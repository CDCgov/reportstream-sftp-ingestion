package utils

// The name of the Azure blob storage container. In future, there will be different ones per customer
const ContainerName = "sftp"

// HL7 messages (NO zips!) placed in this folder trigger a queue message.
// We read the message and send it to ReportStream
const MessageStartingFolderPath = "import"

// HL7 messages are moved from the `MessageStartingFolderPath` to the `SuccessFolder` after
// we receive a success response from ReportStream
const SuccessFolder = "success"

// HL7 messages are moved from the `MessageStartingFolderPath` to the `FailureFolder` after
// we receive a failure response from ReportStream
const FailureFolder = "failure"

// In read_and_send, move files to the `FailureFolder` when we get the below response from ReportStream
const ReportStreamNonTransientFailure = "400"

// Use this when logging an error.
// E.g. `slog.Warn("Failed to construct the ReportStream senders", slog.Any(utils.ErrorKey, err))`
const ErrorKey = "error"
