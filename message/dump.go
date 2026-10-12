package message

import (
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/nqd2/igo8583/field"
)

/* ============================================================================
 * Data Masking & Diagnostics
 * ============================================================================ */

/*
MaskPAN masks a card Primary Account Number (PAN) for PCI-DSS compliance display.
Preserves the first 6 digits (BIN) and last 4 digits, replacing intermediate digits with '*'.
*/
func MaskPAN(pan string) string {
	panLen := len(pan)
	if panLen <= 10 {
		return strings.Repeat("*", panLen)
	}
	first6 := pan[:6]
	last4 := pan[panLen-4:]
	middle := strings.Repeat("*", panLen-10)
	return first6 + middle + last4
}

/*
Dump formats the message into a human-readable XML-like log string.
If maskSensitive is true, sensitive data elements (PAN, tracks, PIN block) are masked.
*/
func (m *Message) Dump(maskSensitive bool) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var sb strings.Builder
	sb.WriteString("<isomsg>\n")

	if len(m.header) > 0 {
		sb.WriteString(fmt.Sprintf("  <!-- header --> <header hex=\"%s\"/>\n", strings.ToUpper(hex.EncodeToString(m.header))))
	}

	var fieldIDs []int
	for id := range m.fields {
		fieldIDs = append(fieldIDs, id)
	}
	sort.Ints(fieldIDs)

	for _, id := range fieldIDs {
		comp := m.fields[id]
		dumpComponent(&sb, id, comp, "  ", maskSensitive)
	}

	sb.WriteString("</isomsg>\n")
	return sb.String()
}

func dumpComponent(sb *strings.Builder, id int, comp *field.Component, indent string, maskSensitive bool) {
	if comp.Subfields != nil && len(comp.Subfields) > 0 {
		sb.WriteString(fmt.Sprintf("%s<!-- %03d --> <composite id=\"%d\">\n", indent, id, id))
		var subIDs []int
		for sID := range comp.Subfields {
			subIDs = append(subIDs, sID)
		}
		sort.Ints(subIDs)
		for _, sID := range subIDs {
			sub := comp.Subfields[sID]
			dumpComponent(sb, sID, sub, indent+"  ", maskSensitive)
		}
		sb.WriteString(fmt.Sprintf("%s</composite>\n", indent))
		return
	}

	if comp.TLV != nil {
		sb.WriteString(fmt.Sprintf("%s<!-- %03d --> <tlv id=\"%d\">\n", indent, id, id))
		for _, node := range comp.TLV.Nodes() {
			val := node.ValueHex()
			if maskSensitive && (node.TagHex() == "5A" || node.TagHex() == "57") {
				val = MaskPAN(val)
			}
			sb.WriteString(fmt.Sprintf("%s  <tag id=\"%s\" value=\"%s\"/>\n", indent, node.TagHex(), val))
		}
		sb.WriteString(fmt.Sprintf("%s</tlv>\n", indent))
		return
	}

	val := comp.GetString()
	if maskSensitive {
		val = applyMasking(id, val, comp.GetBytes())
	}

	sb.WriteString(fmt.Sprintf("%s<!-- %03d --> <field id=\"%d\" value=\"%s\"/>\n", indent, id, id, val))
}

func applyMasking(id int, val string, _ []byte) string {
	switch id {
	case 2:
		return MaskPAN(val)
	case 35, 45:
		if len(val) >= 16 {
			return MaskPAN(val[:16]) + "..."
		}
		return strings.Repeat("*", len(val))
	case 52, 53:
		return "****************"
	default:
		return val
	}
}
