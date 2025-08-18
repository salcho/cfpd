package main

import (
	"fmt"
	"main/messages"
)

type CFDPEntity struct {
	ID      uint
	Name    string
	Service *CFPDService
}

func (c CFDPEntity) GetID() uint {
	return c.ID
}

func (c CFDPEntity) GetName() string {
	return c.Name
}

func (c CFDPEntity) ListDirectory(dstEntityID uint, localPath, remotePath string) error {
	listingRequest := RequestPrimitive{
		ReqType:          PutRequest,
		DstEntityID:      dstEntityID,
		SrcEntityID:      c.ID,
		TransmissionMode: messages.Unacknowledged,
		MessagesToUser: []messages.Message{
			messages.NewDirectoryListingRequest("/path/to/directory", "/path/to/directory/listing.txt"),
		},
	}

	err := c.Service.Request(listingRequest, 1)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return err
	}

	// fakeListing := "file1.txt file2.txt file3.txt"
	// // Create a FileDirective PDU to send metadata about the directory listing
	// pduHeader := messages.NewPDUHeader(false, c.Service.config.entityID, 0, 12345)
	// pduContents := messages.MetadataPDUContents{
	// 	ClosureRequested:    true,
	// 	ChecksumType:        0x9, // Assuming no checksum for simplicity
	// 	FileSize:            uint64(len(fakeListing)),
	// 	SourceFileName:      "/dir/to/list",
	// 	DestinationFileName: "/path/to/respond",
	// }

	// err := c.Service.RequestPDU(pduContents.ToBytes(pduHeader), dstEntityID)
	// if err != nil {
	// 	fmt.Println("Error sending request:", err)
	// 	return err
	// }
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

func NewEntity(id uint, name string, service *CFPDService) CFDPEntity {
	go service.BindAndListen()
	return CFDPEntity{
		ID:      id,
		Name:    name,
		Service: service,
	}
}
