package rest

// MetadataHeaderPrefix is the http prefix that represents custom metadata
// parameters to or from an RPC call.
const MetadataHeaderPrefix = "Yggdrasil-Metadata-"

// MetadataTrailerPrefix is prepended to RPC metadata as it is converted to
// HTTP headers in a response handled by rest
const MetadataTrailerPrefix = "Yggdrasil-Trailer-"
