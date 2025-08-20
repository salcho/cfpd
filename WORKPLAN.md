# CFDP Application Refactoring and Improvement Plan

This document outlines the high-level plan for refactoring and improving the CFDP application. The primary goals are to enhance code quality, readability, testability, and maintainability without altering the existing core functionality.

## Action Items

### 1. Modularize the `messages` Package
- [ ] Break down the monolithic `messages.go` file into smaller, more focused files.
  - [ ] `pdu.go`: For the core `ProtocolDataUnitHeader` and related structures.
  - [ ] `file_directives.go`: For file directive PDUs (e.g., `MetadataPDU`, `EOF`, `Finished`).
  - [ ] `file_data.go`: For the `FileDataPDU`.
  - [ ] `user_messages.go`: For user-defined messages and TLVs.
  - [ ] `checksum.go`: Keep the CRC logic separate.

### 2. Introduce a Transaction Concept
- [ ] Create a `Transaction` struct/object to manage the lifecycle of a single file transfer.
- [ ] The `Transaction` will be responsible for:
  - [ ] Managing the state of the transaction (e.g., sending file data, waiting for ACK).
  - [ ] Sequencing PDUs for the transfer.
  - [ ] Handling timeouts and retransmissions.
  - [ ] Tracking progress.
- [ ] The main `Entity` will manage a map of active transactions, identified by their Transaction ID.

### 3. Refine Configuration Management
- [ ] Enhance the `ServiceConfig` struct to be more comprehensive.
- [ ] Consider loading configuration from a file (e.g., TOML, YAML, or JSON) to avoid hardcoding values.
- [ ] Include settings for timeouts, retransmission limits, and other protocol parameters.

### 4. Enhance Unit Tests
- [ ] Achieve near-complete test coverage for the `messages` package.
- [ ] Add unit tests for all PDU types, focusing on serialization and deserialization edge cases.
- [ ] Mock dependencies (e.g., network, filesystem) to create isolated and reliable tests for higher-level components like the `Transaction` manager.

### 5. Implement Integration Tests
- [ ] Develop end-to-end integration tests that simulate a full file transfer between two CFDP entities running in the same process.
- [ ] These tests should:
  - [ ] Initiate a "put" request from a client entity.
  - [ ] Verify that the server entity receives the file correctly.
  - [ ] Test both acknowledged and unacknowledged transmission modes.
  - [ ] Verify the integrity of the transferred file using checksums.
