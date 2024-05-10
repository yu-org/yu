package types

type Event struct {
	//Caller      *Address `json:"caller"`
	//BlockStage  string   `json:"block_stage"`
	//BlockHash   Hash     `json:"block_hash"`
	//Height      BlockNum `json:"height"`
	//TripodName  string   `json:"tripod_name"`
	//WritingName string   `json:"writing_name"`
	//LeiCost     uint64   `json:"lei_cost"`
	Value []byte `json:"value"`
}

//func (e *Event) DecodeJsonValue(v any) error {
//	return json.Unmarshal(e.Value, v)
//}

//func (e *Event) Sprint() (str string) {
//	if e.BlockStage == ExecuteTxnsStage {
//		str = fmt.Sprintf(
//			"[Event] Caller(%s) call Tripod(%s) Writing(%s) in Block(%s) on Height(%d): %s",
//			e.Caller.String(),
//			e.TripodName,
//			e.WritingName,
//			e.BlockHash.String(),
//			e.Height,
//			e.Value,
//		)
//	} else {
//		str = fmt.Sprintf(
//			"[Event] %s Block(%s) on Height(%d) in Tripod(%s): %s",
//			e.BlockStage,
//			e.BlockHash.String(),
//			e.Height,
//			e.TripodName,
//			e.Value,
//		)
//	}
//	return
//}

//type Events []Event
//
//func ToEvents(events []Event) Events {
//	var es Events
//	es = append(es, events...)
//	return es
//}
//
//func (es Events) ToArray() []Event {
//	return es[:]
//}
//
//func (es Events) Encode() ([]byte, error) {
//	return GobEncode(es)
//}
