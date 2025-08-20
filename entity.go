package main

import (
	"fmt"
	"log/slog"
	filesystem "main/filesystem"
	"main/messages"
	"main/statemachine"
)

type CFDPEntity struct {
	ID      uint16
	Name    string
	service *CFPDService
	fs      filesystem.FS
	sm      *statemachine.StateMachine
}

func (c CFDPEntity) GetID() uint16 {
	return c.ID
}

func (c CFDPEntity) GetName() string {
	return c.Name
}

type PutParameters struct {
	DstEntityID           *uint16
	SrcFileName           string
	DestinationFileName   string
	SegmentationControl   bool
	FaultHandlerOverrides map[messages.ConditionCode]messages.Action
	FlowLabel             string
	TransmissionMode      messages.TransmissionMode
	ClosureRequested      bool
	MessagesToUser        []messages.Message
	FilestoreRequests     []string
}

// The Put.request primitive shall be used by the application to request delivery of a file from the source filestore to a destination filestore.
func (c CFDPEntity) PutRequest(p PutParameters) error {
	if p.DstEntityID == nil {
		fmt.Println("Destination Entity ID is required")
		return fmt.Errorf("destination entity ID is required")
	}

	// check that all conditions in faultHandlerOverrides are error conditions
	if p.FaultHandlerOverrides != nil {
		for condition := range p.FaultHandlerOverrides {
			if !condition.IsError() {
				fmt.Println("Fault handler overrides must only contain error conditions")
				return fmt.Errorf("fault handler overrides must only contain error conditions")
			}
		}
	}

	// TODO: select a default error handler if none is provided, according to the contents of the MIB
	// TODO: implement flow label handling
	// TODO: implement closure requested handling

	dstID := *p.DstEntityID
	slog.Debug("Triggering Put.request", "from", c.ID, "to", dstID)
	for _, msg := range p.MessagesToUser {
		switch msg.GetMessageType() {
		case messages.MessageTypeDirectoryRequest:
			listingRequest := msg.(*messages.DirectoryListingRequest)
			slog.Debug("Sending Directory Request", "remote", listingRequest.DirToList, "local", listingRequest.PathToRespond, "from", c.ID, "dst", dstID)

			// TODO: assign sequence number and transaction ID properly
			pdu, err := messages.NewMetadataPDU(false, c.ID, dstID, 12345, messages.FileDirective, p.ClosureRequested,
				listingRequest.PathToRespond, listingRequest.DirToList, []messages.Message{listingRequest})
			if err != nil {
				fmt.Println("Error creating File Directive PDU:", err)
				return err
			}
			if err := c.service.RequestBytes(pdu.ToBytes(), dstID); err != nil {
				fmt.Println("Error sending PDU:", err)
			}
			c.sm.SetState(statemachine.WaitingForDirectoryListing)
		default:
			fmt.Println(c.ID, "Unknown request primitive: ", msg.GetMessageType())
			return fmt.Errorf("unknown request primitive: %s", msg.GetMessageType())
		}
	}
	return nil
}

func (c CFDPEntity) HandlePDU(pdu messages.PDU) error {
	if pdu.GetType() == messages.FileDirective {
		pdu := pdu.(*messages.FileDirectivePDU)
		switch pdu.DirCode {
		case messages.MetadataPDU:
			slog.Debug("Handling Metadata PDU", "entityID", c.ID)
			metadata := messages.MetadataPDUContents{}
			err := metadata.FromBytes(pdu.Data, pdu.Header)
			if err != nil {
				return fmt.Errorf("failed to decode metadata: %v", err)
			}

			if hasUploadRequest(metadata.MessagesToUser) {
				c.sm.SetState(statemachine.SendingDirectoryListing)
				slog.Debug("Handling Directory Listing Request", "entityID", c.ID, "remotePath", metadata.SourceFileName, "localPath", metadata.DestinationFileName)
				return c.UploadDirectoryListing(pdu, metadata)
			}

			c.sm.SetState(statemachine.WaitingForFileData)
			c.sm.Context.FilePath = metadata.DestinationFileName
			return nil

		default:
			return fmt.Errorf("unknown directive code: %v", pdu.DirCode)
		}
	}

	if pdu.GetType() == messages.FileData {
		pdu := pdu.(*messages.FileDataPDU)
		slog.Debug("Handling File Data PDU", "entityID", c.ID, "dataLength", len(pdu.FileData))
		if c.sm.CurrentState != statemachine.WaitingForFileData {
			return fmt.Errorf("unexpected state: %v, expected WaitingForFileData", c.sm.CurrentState)
		}

		err := c.fs.WriteFile(c.sm.Context.FilePath, pdu.FileData)
		if err != nil {
			return fmt.Errorf("failed to write file data: %v", err)
		}
		slog.Info("File data written successfully", "fileName", c.sm.Context.FilePath, "entityID", c.ID)
		return nil
	}

	return nil
}

func hasUploadRequest(m []messages.Message) bool {
	for _, msg := range m {
		if msg.GetMessageType() == messages.MessageTypeDirectoryRequest {
			return true
		}
	}
	return false
}

func (c CFDPEntity) UploadDirectoryListing(pdu *messages.FileDirectivePDU, m messages.MetadataPDUContents) error {
	slog.Info("Uploading directory listing", "local", m.DestinationFileName, "uploadTo", m.SourceFileName, "from", c.ID, "to", pdu.Header.SourceEntityID)

	l, err := c.fs.ListDirectory(m.DestinationFileName)
	if err != nil {
		return fmt.Errorf("failed to list directory: %v", err)
	}

	return c.CopyOperation(pdu.Header.SourceEntityID, m.DestinationFileName, m.SourceFileName, l)
}

func (c CFDPEntity) CopyOperation(dstEntityID uint16, srcFileName, dstFileName string, contents string) error {
	slog.Debug("Copying file", "from", srcFileName, "to", dstFileName, "dstEntityID", dstEntityID, "entityID", c.ID)

	pdu, err := messages.NewMetadataPDU(false, c.ID, dstEntityID, 12345, messages.FileDirective, true,
		srcFileName, dstFileName, []messages.Message{
			&messages.DirectoryListingResponse{
				WasSuccessful: true,
				DirToList:     dstFileName,
				PathToRespond: srcFileName,
			},
		})
	if err != nil {
		fmt.Println("Error creating File Directive PDU:", err)
		return nil
	}

	if err := c.service.RequestBytes(pdu.ToBytes(), dstEntityID); err != nil {
		fmt.Println("Error sending PDU:", err)
		return err
	}

	fdp := messages.FileDataPDU{
		Header:   messages.NewPDUHeader(false, c.ID, dstEntityID, 12345, messages.FileData),
		FileData: []byte(contents),
		Offset:   0,
	}

	if err := c.service.RequestBytes(fdp.ToBytes(), dstEntityID); err != nil {
		fmt.Println("Error sending File Data PDU:", err)
		return err
	}

	slog.Info("File data sent successfully", "fileName", dstFileName, "entityID", c.ID, "dstEntityID", dstEntityID)
	return nil
}

func NewEntity(id uint16, name string, sc ServiceConfig) CFDPEntity {
	service := &CFPDService{Config: sc}
	entity := CFDPEntity{
		ID:      id,
		Name:    name,
		service: service,
		fs:      &filesystem.LocalFS{},
		sm:      statemachine.NewStateMachine(),
	}
	service.Bind(&entity)
	return entity
}
