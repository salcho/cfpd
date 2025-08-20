package main

import (
	"log/slog"
	"cfdp/messages"
)

type CFDPApp struct {
	Entity *CFDPEntity
}

func NewCFDPApp(id uint16, name string, sc ServiceConfig) *CFDPApp {
	app := &CFDPApp{Entity: NewEntity(id, name, sc)}
	app.Entity.SetIndicationCallback(func(i Indication) {
		switch i {
		case Transaction:
			slog.Info("Received indication:", "indication", "Transaction started")
			return
		case EOFSent:
			slog.Info("Received indication:", "indication", "EOF sent")
			return
		case TransactionFinished:
			slog.Info("Received indication:", "indication", "Transaction finished")
			return
		case MetadataRecv:
			slog.Info("Received indication:", "indication", "Metadata received")
			return
		case FileSegmentRecv:
			slog.Info("Received indication:", "indication", "File segment received")
			return
		case Report:
			slog.Info("Received indication:", "indication", "Report received")
			return
		case Suspended:
			slog.Info("Received indication:", "indication", "Service suspended")
			return
		case Resumed:
			slog.Info("Received indication:", "indication", "Service resumed")
			return
		case Fault:
			slog.Info("Received indication:", "indication", "Fault occurred")
			return
		case Abandoned:
			slog.Info("Received indication:", "indication", "Transaction abandoned")
			return
		case EOFRecv:
			slog.Info("Received indication:", "indication", "EOF received")
			return
		default:
			slog.Info("Received indication:", "indication", "Unknown indication")
			return
		}
	})
	return app
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
