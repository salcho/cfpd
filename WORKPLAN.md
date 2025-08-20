# CFDP Application Refactoring and Improvement Plan

This document outlines the high-level plan for refactoring and improving the CFDP application. The primary goals are to enhance code quality, readability, testability, and maintainability without altering the existing core functionality.

## Action Items

### 1. Modularize the `messages` Package
- [x] Break down the monolithic `messages.go` file into smaller, more focused files.
  - [x] `pdu.go`: For the core `ProtocolDataUnitHeader` and related structures.
  - [x] `file_directives.go`: For file directive PDUs (e.g., `MetadataPDU`, `EOF`, `Finished`).
  - [x] `file_data.go`: For the `FileDataPDU`.
  - [x] `user_messages.go`: For user-defined messages and TLVs.
  - [x] `checksum.go`: Keep the CRC logic separate.

### 2. Enhance Unit Tests
- [ ] Ensure all existing tests are passing.
- [ ] Achieve near-complete test coverage for the `entity`, `service` and `app` files.

### 3. Implement Integration Tests
- [ ] Develop end-to-end integration tests that simulate a full file transfer between two CFDP entities running in the same process.
- [ ] These tests should:
  - [ ] Initiate a "put" request from a client entity.
  - [ ] Verify that the server entity receives the file correctly.
  - [ ] Test both acknowledged and unacknowledged transmission modes.
  - [ ] Verify the integrity of the transferred file using checksums.
