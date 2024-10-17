package callqueuepolicy

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PolicyType int

const (
	VoiceMail PolicyType = iota
)

func (pt PolicyType) String() string {
	switch pt {
	case VoiceMail:
		return "voice_mail"
	default:
		return ""
	}
}

type readDto struct {
	callQueueID            types.String
	policyVoiceMailMembers []*readDtoPolicyVoiceMailMember
}

type readDtoPolicyVoiceMailMember struct {
	accessUserID  types.String
	allowDownload types.Bool
	allowDelete   types.Bool
	allowSharing  types.Bool
	sharedID      types.String
}

type addDto struct {
	callQueueID            types.String
	policyType             PolicyType
	voicemailAccessMembers []*addDtoVoicemailAccessMember
}

type addDtoVoicemailAccessMember struct {
	accessUserID  types.String
	allowDownload types.Bool
	allowDelete   types.Bool
	allowSharing  types.Bool
}

type updateDto struct {
	callQueueID            types.String
	policyType             PolicyType
	voicemailAccessMembers []*updateDtoVoicemailAccessMember
}

type updateDtoVoicemailAccessMember struct {
	accessUserID  types.String
	allowDownload types.Bool
	allowDelete   types.Bool
	allowSharing  types.Bool
	sharedID      types.String
}

type removeDto struct {
	callQueueID types.String
	policyType  PolicyType
	sharedIDs   []types.String
}
