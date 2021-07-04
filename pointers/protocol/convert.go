package protocol

func Operate2StoreValue(op *Operate) *StoreValue {
	return &StoreValue{
		Value: *op.Value,
		Sign:  op.Sign,
		Time:  op.Time,
	}
}

func StoreValue2Operate(val *StoreValue) *Operate {
	return &Operate{
		Value: &val.Value,
		Sign:  val.Sign,
		Time:  val.Time,
	}
}

func Key2Topic(key string) string {
	return TopicPrefix + key
}
