package main

import (
	"fmt"
	"log/slog"
	filesystem "main/filesystem"
	"main/messages"
)

type CFDPEntity struct {
	ID      uint16
	Name    string
	service *CFPDService
	fs      filesystem.FS
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
			pdu, err := messages.NewFileDirectivePDU(false, c.ID, dstID, 12345, messages.FileDirective, p.ClosureRequested,
				listingRequest.PathToRespond, listingRequest.DirToList, []messages.Message{listingRequest})
			if err != nil {
				fmt.Println("Error creating File Directive PDU:", err)
				return err
			}
			if err := c.service.RequestBytes(pdu.ToBytes(int16(len(pdu.Data))), dstID); err != nil {
				fmt.Println("Error sending PDU:", err)
			}
		case messages.MessageTypeDirectoryResponse:
			listingResponse := msg.(*messages.DirectoryListingResponse)
			slog.Debug("Sending Directory Response", "remote", listingResponse.PathToRespond, "local", listingResponse.DirToList, "from", c.ID, "dst", dstID)
		default:
			fmt.Println(c.ID, "Unknown request primitive: ", msg.GetMessageType())
			return fmt.Errorf("unknown request primitive: %s", msg.GetMessageType())
		}
	}
	return nil
}

func (c CFDPEntity) ProxyOperation() {
	fmt.Println("Performing proxy operation for CFDP entity:", c.Name)
}

func (c CFDPEntity) StatusReportOperation() {
	fmt.Println("Performing status report operation for CFDP entity:", c.Name)
}

func (c CFDPEntity) SuspendOperation() {
	fmt.Println("Suspending CFDP entity:", c.Name)
}

func (c CFDPEntity) ResumeOperation() {
	fmt.Println("Resuming CFDP entity:", c.Name)
}

func (c CFDPEntity) HandlePDU(pdu messages.FileDirectivePDU) error {
	switch pdu.DirCode {
	case messages.MetadataPDU:
		slog.Debug("Handling Metadata PDU", "entityID", c.ID)
		metadata := messages.MetadataPDUContents{}
		err := metadata.FromBytes(pdu.Data, pdu.Header)
		if err != nil {
			return fmt.Errorf("failed to decode metadata: %v", err)
		}
		return c.UploadDirectoryListing(pdu, metadata)

	default:
		return fmt.Errorf("unknown directive code: %v", pdu.DirCode)
	}
}

func (c CFDPEntity) UploadDirectoryListing(pdu messages.FileDirectivePDU, m messages.MetadataPDUContents) error {
	slog.Info("Uploading directory listing", "local", m.DestinationFileName, "uploadTo", m.SourceFileName, "localEntity", c.ID, "remoteEntity", pdu.Header.SourceEntityID)

	_, err := c.fs.ListDirectory(m.DestinationFileName)
	if err != nil {
		return fmt.Errorf("failed to list directory: %v", err)
	}

	c.PutRequest(PutParameters{
		DstEntityID:           &pdu.Header.SourceEntityID,
		SrcFileName:           m.DestinationFileName,
		DestinationFileName:   m.SourceFileName,
		SegmentationControl:   false,
		FaultHandlerOverrides: make(map[messages.ConditionCode]messages.Action),
		FlowLabel:             "",
		TransmissionMode:      messages.Unacknowledged,
		ClosureRequested:      true,
		MessagesToUser: []messages.Message{
			&messages.DirectoryListingResponse{
				WasSuccessful: true,
				DirToList:     m.DestinationFileName,
				PathToRespond: m.SourceFileName,
			},
			// TODO: use a proper transaction ID
			&messages.OriginatingTransactionID{
				SourceEntityID:            c.ID,
				TransactionSequenceNumber: 12345,
			},
		},
		FilestoreRequests: []string{},
	})

	return nil
}

func NewEntity(id uint16, name string, sc ServiceConfig) CFDPEntity {
	service := &CFPDService{Config: sc}
	entity := CFDPEntity{
		ID:      id,
		Name:    name,
		service: service,
		fs:      &filesystem.LocalFS{},
	}
	service.Bind(&entity)
	return entity
}
