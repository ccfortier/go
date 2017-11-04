package caste

import (
	"fmt"
)

type CasteMsg struct {
	PId  int
	CId  int
	Text string
}

func (cm CasteMsg) Encode() string {
	return fmt.Sprintf("%v", cm)
}

func (cm *CasteMsg) Decode(cmEncoded string) {
	fmt.Sscanf(cmEncoded[:len(cmEncoded)-1], "{%d %d %s", &cm.PId, &cm.CId, &cm.Text)
}
