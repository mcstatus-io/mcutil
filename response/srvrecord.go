package response

type SRVRecord struct {
	Host string `json:"host"`
	Port uint16 `json:"port"`
}
