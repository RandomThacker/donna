package constant

// Conversation / message public ID prefixes and enums.
const (
	PublicIDPrefixConversation = "conv_"
	PublicIDPrefixMessage      = "msg_"

	ConversationChannelWeb = "web"
	ConversationStatusActive = "active"
	ConversationPurposeGeneral = "general"

	MessageRoleUser      = "user"
	MessageRoleAssistant = "assistant"
	MessageRoleSystem    = "system"

	MessageContentFormatPlain = "plain"

	// ChatHistoryDefaultLimit caps messages returned for the primary thread.
	ChatHistoryDefaultLimit = 200
)
