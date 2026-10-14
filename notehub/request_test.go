package notehub

import (
	"encoding/json"
	"testing"

	"github.com/blues/note-go/note"
	"github.com/stretchr/testify/require"
)

func TestSkillUploadRequest(t *testing.T) {
	require.Equal(t, UploadTypeSkill, ParseUploadType("skill"))
	require.Equal(t, UploadTypeUserData, ParseUploadType("data"))

	for _, operation := range []string{HubAppUploadRead, HubAppUploads} {
		t.Run(operation, func(t *testing.T) {
			var req HubRequest
			require.NoError(t, note.JSONUnmarshal([]byte(`{"req":"`+operation+`","type":"skill","full":true}`), &req))
			require.Equal(t, UploadTypeSkill, req.FileType())
			require.True(t, req.Full)
			encoded, err := note.JSONMarshal(req)
			require.NoError(t, err)
			require.JSONEq(t, `{"req":"`+operation+`","type":"skill","full":true}`, string(encoded))
		})
	}
}

func TestUploadMetadataPayloadJSON(t *testing.T) {
	for _, tc := range []struct {
		name    string
		payload []byte
		full    bool
		want    string
	}{
		{"metadata only", nil, false, `{"name":"bundle.zip"}`},
		{"binary", []byte{0, 1, 2, 0xff}, true, `{"name":"bundle.zip","payload":"AAEC/w=="}`},
		{"empty file", []byte{}, true, `{"name":"bundle.zip","payload":""}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			metadata := UploadMetadata{Name: "bundle.zip"}
			if tc.full {
				metadata.SetContents(tc.payload)
			}
			// Both the standard encoder and Notehub's encoder must use base64.
			for _, marshal := range []func(any) ([]byte, error){json.Marshal, note.JSONMarshal} {
				encoded, err := marshal(metadata)
				require.NoError(t, err)
				require.JSONEq(t, tc.want, string(encoded))
				var decoded UploadMetadata
				require.NoError(t, note.JSONUnmarshal(encoded, &decoded))
				require.Equal(t, metadata, decoded)
			}
		})
	}
}

func TestUploadMetadataSkillTextJSON(t *testing.T) {
	for _, tc := range []struct {
		name string
		text string
		full bool
		want string
	}{
		{"metadata only", "", false, `{"name":"SKILL.md","type":"skill"}`},
		{"text", "# Skill\nUse café, 日本語, 🌍, and \"quotes\".\n", true, `{"name":"SKILL.md","type":"skill","text":"# Skill\nUse café, 日本語, 🌍, and \"quotes\".\n"}`},
		{"empty file", "", true, `{"name":"SKILL.md","type":"skill","text":""}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			metadata := UploadMetadata{Name: "SKILL.md", FileType: UploadTypeSkill}
			if tc.full {
				metadata.SetContents([]byte(tc.text))
			}
			for _, marshal := range []func(any) ([]byte, error){json.Marshal, note.JSONMarshal} {
				encoded, err := marshal(metadata)
				require.NoError(t, err)
				require.JSONEq(t, tc.want, string(encoded))
				var decoded UploadMetadata
				require.NoError(t, note.JSONUnmarshal(encoded, &decoded))
				require.Equal(t, metadata, decoded)
			}
		})
	}
}
