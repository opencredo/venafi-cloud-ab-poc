package secretfixer

import (
	"bytes"
	"encoding/pem"
)

func caFromChain(chain []byte) ([]byte, error) {
	var certs []*pem.Block
	for {
		var pemBlock *pem.Block
		pemBlock, chain = pem.Decode(chain)
		if pemBlock == nil {
			break
		}

		certs = append(certs, pemBlock)
	}

	if len(certs) < 2 {
		return []byte{}, nil
	}

	byteBuf := bytes.Buffer{}
	for _, c := range certs[1:] {
		err := pem.Encode(&byteBuf, c)
		if err != nil {
			return []byte{}, err
		}
	}

	return byteBuf.Bytes(), nil
}
