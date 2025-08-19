package main

import (
	"log/slog"
	"main/messages"
)

type CFDPApp struct {
	Entity *CFDPEntity
}

func (app *CFDPApp) ListDirectory(dstEntityID uint16, localPath, remotePath string) error {
	slog.Info("Listing remote directory", "remotePath", remotePath, "remoteEntity", dstEntityID, "localPath", localPath, "localEntity", app.Entity.ID)

	app.Entity.PutRequest(PutParameters{
		DstEntityID:           &dstEntityID,
		SrcFileName:           localPath,
		DestinationFileName:   remotePath,
		SegmentationControl:   false,
		FaultHandlerOverrides: make(map[messages.ConditionCode]messages.Action),
		FlowLabel:             "",
		TransmissionMode:      messages.Unacknowledged,
		ClosureRequested:      true,
		MessagesToUser: []messages.Message{
			&messages.DirectoryListingRequest{
				DirToList:     remotePath,
				PathToRespond: localPath,
			},
		},
		FilestoreRequests: []string{},
	})

	return nil
}
