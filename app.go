package main

import (
	"fmt"
	"main/messages"
)

type CFDPApp interface {
	ListDirectory(dstEntityID uint16, localPath, remotePath string) error
}

type cfpdApp struct {
	entity *CFDPEntity
}

func NewCFDPApp(e *CFDPEntity) CFDPApp {
	if e == nil {
		panic("CFDPApp cannot be created with a nil CFDPEntity")
	}
	return &cfpdApp{e}
}

func (app *cfpdApp) ListDirectory(dstEntityID uint16, localPath, remotePath string) error {
	fmt.Println("Listing directory ", remotePath, "@", dstEntityID, ", will store in local path", localPath, "@", app.entity.ID)

	app.entity.PutRequest(PutParameters{
		DstEntityID:           &dstEntityID,
		SrcFileName:           localPath,
		DestinationFileName:   remotePath,
		SegmentationControl:   false,
		FaultHandlerOverrides: make(map[messages.ConditionCode]messages.Action),
		FlowLabel:             "",
		TransmissionMode:      messages.Unacknowledged,
		ClosureRequested:      true,
		MessagesToUser: []messages.Message{
			messages.NewDirectoryListingRequest(localPath, remotePath),
		},
		FilestoreRequests: []string{},
	})

	return nil
}
