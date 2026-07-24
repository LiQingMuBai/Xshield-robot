package fixedfloat

func MapToResponse(m map[string]interface{}) (*Response, error) {
	resp := &Response{}

	code, ok := getFloat64(m, "code")
	if !ok {
		return nil, errInvalidFixedFloatField("code")
	}
	msg, ok := getString(m, "msg")
	if !ok {
		return nil, errInvalidFixedFloatField("msg")
	}
	dataMap, ok := getMap(m, "data")
	if !ok {
		return nil, errInvalidFixedFloatField("data")
	}
	data, err := mapToData(dataMap)
	if err != nil {
		return nil, err
	}

	resp.Code = code
	resp.Msg = msg
	resp.Data = data

	return resp, nil
}

func mapToData(m map[string]interface{}) (Data, error) {
	d := Data{}

	backMap, ok := getMap(m, "back")
	if !ok {
		return Data{}, errInvalidFixedFloatField("data.back")
	}
	back, err := mapToAddressInfo(backMap)
	if err != nil {
		return Data{}, err
	}

	fromMap, ok := getMap(m, "from")
	if !ok {
		return Data{}, errInvalidFixedFloatField("data.from")
	}
	from, err := mapToAddressInfo(fromMap)
	if err != nil {
		return Data{}, err
	}

	timeMap, ok := getMap(m, "time")
	if !ok {
		return Data{}, errInvalidFixedFloatField("data.time")
	}
	timeInfo, err := mapToTimeInfo(timeMap)
	if err != nil {
		return Data{}, err
	}

	toMap, ok := getMap(m, "to")
	if !ok {
		return Data{}, errInvalidFixedFloatField("data.to")
	}
	to, err := mapToAddressInfo(toMap)
	if err != nil {
		return Data{}, err
	}

	id, ok := getString(m, "id")
	if !ok {
		return Data{}, errInvalidFixedFloatField("data.id")
	}
	status, ok := getString(m, "status")
	if !ok {
		return Data{}, errInvalidFixedFloatField("data.status")
	}
	token, ok := getString(m, "token")
	if !ok {
		return Data{}, errInvalidFixedFloatField("data.token")
	}
	orderType, ok := getString(m, "type")
	if !ok {
		return Data{}, errInvalidFixedFloatField("data.type")
	}

	d.Back = back
	d.Email = m["email"]
	d.From = from
	d.ID = id
	d.Status = status
	d.Time = timeInfo
	d.To = to
	d.Token = token
	d.Type = orderType

	return d, nil
}

func mapToEmergency(m map[string]interface{}) (Emergency, error) {
	e := Emergency{}

	choice, ok := getString(m, "choice")
	if !ok {
		return Emergency{}, errInvalidFixedFloatField("emergency.choice")
	}
	repeat, ok := getInt(m, "repeat")
	if !ok {
		return Emergency{}, errInvalidFixedFloatField("emergency.repeat")
	}
	status, ok := m["status"].([]interface{})
	if !ok {
		return Emergency{}, errInvalidFixedFloatField("emergency.status")
	}

	e.Choice = choice
	e.Repeat = repeat
	e.Status = status
	return e, nil
}

func mapToTimeInfo(m map[string]interface{}) (TimeInfo, error) {
	t := TimeInfo{}

	expiration, ok := getFloat64(m, "expiration")
	if !ok {
		return TimeInfo{}, errInvalidFixedFloatField("time.expiration")
	}
	left, ok := getFloat64(m, "left")
	if !ok {
		return TimeInfo{}, errInvalidFixedFloatField("time.left")
	}
	reg, ok := getFloat64(m, "reg")
	if !ok {
		return TimeInfo{}, errInvalidFixedFloatField("time.reg")
	}
	update, ok := getFloat64(m, "update")
	if !ok {
		return TimeInfo{}, errInvalidFixedFloatField("time.update")
	}

	t.Expiration = expiration
	t.Left = left
	t.Reg = reg
	t.Update = update

	if f, ok := getFloat64(m, "finish"); ok {
		val := f
		t.Finish = &val
	}
	if s, ok := getFloat64(m, "start"); ok {
		val := s
		t.Start = &val
	}

	return t, nil
}

func ExtractTime(input map[string]interface{}) (TimeInfo, bool) {
	data, ok := getMap(input, "data")
	if !ok {
		return TimeInfo{}, false
	}

	timeMap, ok := getMap(data, "time")
	if !ok {
		return TimeInfo{}, false
	}

	ti, err := mapToTimeInfo(timeMap)
	if err != nil {
		return TimeInfo{}, false
	}

	return ti, true
}

func ExtractFromAndTo(input map[string]interface{}) (from, to AddressInfo, ok bool) {
	data, exists := getMap(input, "data")
	if !exists {
		return AddressInfo{}, AddressInfo{}, false
	}

	fromMap, fromExists := getMap(data, "from")
	toMap, toExists := getMap(data, "to")
	if !fromExists || !toExists {
		return AddressInfo{}, AddressInfo{}, false
	}

	from, err := mapToAddressInfo(fromMap)
	if err != nil {
		return AddressInfo{}, AddressInfo{}, false
	}
	to, err = mapToAddressInfo(toMap)
	if err != nil {
		return AddressInfo{}, AddressInfo{}, false
	}

	return from, to, true
}

func mapToAddressInfo(m map[string]interface{}) (AddressInfo, error) {
	a := AddressInfo{}

	address, ok := getString(m, "address")
	if !ok {
		return AddressInfo{}, errInvalidFixedFloatField("address.address")
	}
	alias, ok := getString(m, "alias")
	if !ok {
		return AddressInfo{}, errInvalidFixedFloatField("address.alias")
	}
	code, ok := getString(m, "code")
	if !ok {
		return AddressInfo{}, errInvalidFixedFloatField("address.code")
	}
	coin, ok := getString(m, "coin")
	if !ok {
		return AddressInfo{}, errInvalidFixedFloatField("address.coin")
	}
	name, ok := getString(m, "name")
	if !ok {
		return AddressInfo{}, errInvalidFixedFloatField("address.name")
	}
	network, ok := getString(m, "network")
	if !ok {
		return AddressInfo{}, errInvalidFixedFloatField("address.network")
	}
	tag, ok := getString(m, "tag")
	if !ok {
		return AddressInfo{}, errInvalidFixedFloatField("address.tag")
	}
	txMap, ok := getMap(m, "tx")
	if !ok {
		return AddressInfo{}, errInvalidFixedFloatField("address.tx")
	}

	a.Address = address
	a.Alias = alias
	a.Code = code
	a.Coin = coin
	a.Name = name
	a.Network = network
	a.Tag = tag
	a.AddressAlt = m["addressAlt"]
	if amt, ok := getString(m, "amount"); ok {
		a.Amount = &amt
	}
	if mc, ok := getInt(m, "maxConfirmations"); ok {
		a.MaxConfirmations = mc
	}
	if rc, ok := getInt(m, "reqConfirmations"); ok {
		a.ReqConfirmations = rc
	}
	a.TagName = m["tagName"]
	a.Tx = TxInfo{
		Amount:        txMap["amount"],
		CcyFee:        txMap["ccyfee"],
		Confirmations: txMap["confirmations"],
		Fee:           txMap["fee"],
		ID:            txMap["id"],
		TimeBlock:     txMap["timeBlock"],
		TimeReg:       txMap["timeReg"],
	}

	return a, nil
}

func ExtractIDAndStatus(input map[string]interface{}) (id string, status string, ok bool) {
	data, exists := getMap(input, "data")
	if !exists {
		return "", "", false
	}

	idVal, idExists := getString(data, "id")
	statusVal, statusExists := getString(data, "status")

	if !idExists || !statusExists {
		return "", "", false
	}

	return idVal, statusVal, true
}

func getMap(input map[string]interface{}, key string) (map[string]interface{}, bool) {
	value, ok := input[key]
	if !ok || value == nil {
		return nil, false
	}
	result, ok := value.(map[string]interface{})
	return result, ok
}

func getString(input map[string]interface{}, key string) (string, bool) {
	value, ok := input[key]
	if !ok || value == nil {
		return "", false
	}
	result, ok := value.(string)
	return result, ok
}

func getFloat64(input map[string]interface{}, key string) (float64, bool) {
	value, ok := input[key]
	if !ok || value == nil {
		return 0, false
	}
	result, ok := value.(float64)
	return result, ok
}

func getInt(input map[string]interface{}, key string) (int, bool) {
	value, ok := input[key]
	if !ok || value == nil {
		return 0, false
	}

	switch v := value.(type) {
	case int:
		return v, true
	case float64:
		return int(v), true
	default:
		return 0, false
	}
}

func errInvalidFixedFloatField(field string) error {
	return &FieldError{Field: field}
}

type FieldError struct {
	Field string
}

func (e *FieldError) Error() string {
	return "invalid fixedfloat field: " + e.Field
}
