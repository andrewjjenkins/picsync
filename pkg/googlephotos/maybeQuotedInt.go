package googlephotos

import "encoding/json"

// The Google Photos API sometimes returns ints wrapped in quotes.
//
//	{
//		  mediaItemsCount: "6"
//	}
//
// Go's default JSON parser will only allow this as the string "6".  Define
// a new optionally-quoted int type, where the custom unmarshaller will pull
// off one pair of quotes, if present.
type MaybeQuotedInt64 int64

func (i MaybeQuotedInt64) MarshalJSON() ([]byte, error) {
	return json.Marshal(int64(i))
}

func (i *MaybeQuotedInt64) UnmarshalJSON(data []byte) error {
	if len(data) >= 2 && data[0] == '"' && data[len(data)-1] == '"' {
		data = data[1 : len(data)-1]
	}

	var tmp int64
	err := json.Unmarshal(data, &tmp)
	if err != nil {
		return err
	}
	*i = MaybeQuotedInt64(tmp)
	return nil
}
