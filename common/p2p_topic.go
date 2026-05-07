package common

const (
	StartBlockTopic      = "start-block"
	EndBlockTopic        = "end-block"
	FinalizeBlockTopic   = "finalize-block"
	UnpackedWritingTopic = "unpacked-writing"
)

const TopicWritingTopicPrefix = "topic_writing_"

func TopicWritingTopic(topic string) string {
	return TopicWritingTopicPrefix + topic
}
