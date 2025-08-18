package main

import (
	"fmt"
	"log/slog"
	"main/messages"
)

type CFDPEntity struct {
	ID      uint16
	Name    string
	service *CFPDService
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
	slog.Info("Sending Put.request", "dst", dstID, "src", c.ID)
	for _, msg := range p.MessagesToUser {
		switch msg.GetMessageType() {
		case messages.MessageTypeDirectoryRequest:
			listingRequest := msg.(messages.DirectoryListingRequest)
			slog.Info("Sending Directory Request", "dir", listingRequest.DirToList, "file", listingRequest.PathToRespond, "entityID", c.ID)

			// Create a FileDirective PDU to send metadata about the directory listing
			pduHeader := messages.NewPDUHeader(false, c.ID, dstID, 12345)
			fakeListing := []string{"file1.txt", "file2.txt", "file3.txt"}
			pduContents := messages.MetadataPDUContents{
				ClosureRequested:    false,
				ChecksumType:        0, // Assuming no checksum for simplicity
				FileSize:            uint64(len(fakeListing)),
				SourceFileName:      listingRequest.DirToList,
				DestinationFileName: listingRequest.PathToRespond,
			}
			pdu := messages.FileDirectivePDU{
				Header:  pduHeader,
				DirCode: messages.MetadataPDU,
				Data:    pduContents.ToBytes(pduHeader),
			}
			if err := c.service.RequestBytes(pdu.ToBytes(uint16(len(pdu.Data))), c.ID); err != nil {
				fmt.Println("Error sending PDU:", err)
			}

			if err := c.service.Request(
				RequestPrimitive{
					ReqType:          PutRequest,
					DstEntityID:      dstID,
					SrcEntityID:      c.ID,
					TransmissionMode: messages.Acknowledged,
					MessagesToUser: []messages.Message{
						messages.NewDirectoryListingResponse(listingRequest.DirToList, listingRequest.PathToRespond),
						messages.NewOriginatingTransactionID(1, 12345),
					},
				},
				c.ID,
			); err != nil {
				fmt.Println("Error handling directory request:", err)
				return err
			}
		case messages.MessageTypeDirectoryResponse:
			listingResponse := msg.(messages.DirectoryListingResponse)
			slog.Info("Handling Directory Response", "dir", listingResponse.DirToList, "file", listingResponse.PathToRespond, "entityID", c.ID)
			return nil
		case messages.MessageTypeOriginatingTransactionID:
			originatingTransactionID := msg.(messages.OriginatingTransactionID)
			fmt.Println(c.ID, "Handling Originating Transaction ID:", originatingTransactionID.SourceEntityID, originatingTransactionID.TransactionSequenceNumber)
			return nil
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

func NewEntity(id uint16, name string, service *CFPDService) CFDPEntity {
	service.Bind()
	return CFDPEntity{
		ID:      id,
		Name:    name,
		service: service,
	}
}
