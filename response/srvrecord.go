package response

// SRVRecord is the DNS SRV records performed during a lookup.
type SRVRecord struct {
	Host string `json:"host"`
	Port uint16 `json:"port"`
}
