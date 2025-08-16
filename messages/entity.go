package messages

import "fmt"

type CFDPEntity struct {
	ID, Name string
}

func (c CFDPEntity) GetID() string {
	return c.ID
}

func (c CFDPEntity) GetName() string {
	return c.Name
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
