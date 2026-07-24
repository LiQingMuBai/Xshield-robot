package fixedfloat

func MapToResponse(m map[string]interface{}) (*Response, error) {
	resp := &Response{}

	resp.Code = m["code"].(float64)
	resp.Msg = m["msg"].(string)

	dataMap := m["data"].(map[string]interface{})
	resp.Data = mapToData(dataMap)

	return resp, nil
}

func mapToData(m map[string]interface{}) Data {
	d := Data{}

	d.Back = mapToAddressInfo(m["back"].(map[string]interface{}))
	d.Email = m["email"] // 可能是 nil
	d.From = mapToAddressInfo(m["from"].(map[string]interface{}))
	d.ID = m["id"].(string)
	d.Status = m["status"].(string)
	d.Time = mapToTimeInfo(m["time"].(map[string]interface{}))
	d.To = mapToAddressInfo(m["to"].(map[string]interface{}))
	d.Token = m["token"].(string)
	d.Type = m["type"].(string)

	return d
}

func mapToEmergency(m map[string]interface{}) Emergency {
	e := Emergency{}
	e.Choice = m["choice"].(string)
	e.Repeat = m["repeat"].(int)
	e.Status = m["status"].([]interface{})
	return e
}

func mapToTimeInfo(m map[string]interface{}) TimeInfo {
	t := TimeInfo{}

	t.Expiration = m["expiration"].(float64)
	t.Left = m["left"].(float64)
	t.Reg = m["reg"].(float64)
	t.Update = m["update"].(float64)

	// 处理可空字段
	if f, ok := m["finish"]; ok && f != nil {
		val := f.(float64)
		t.Finish = &val
	}
	if s, ok := m["start"]; ok && s != nil {
		val := s.(float64)
		t.Start = &val
	}

	return t
}

func ExtractTime(input map[string]interface{}) (TimeInfo, bool) {
	// 安全访问 data.time
	data, ok := input["data"].(map[string]interface{})
	if !ok {
		return TimeInfo{}, false
	}

	timeMap, ok := data["time"].(map[string]interface{})
	if !ok {
		return TimeInfo{}, false
	}

	var ti TimeInfo

	// 必填字段（根据你的数据，这些都存在）
	ti.Expiration = timeMap["expiration"].(float64)
	ti.Left = timeMap["left"].(float64)
	ti.Reg = timeMap["reg"].(float64)
	ti.Update = timeMap["update"].(float64)

	// 可空字段：Finish 和 Start
	if f := timeMap["finish"]; f != nil {
		val := f.(float64)
		ti.Finish = &val
	}
	if s := timeMap["start"]; s != nil {
		val := s.(float64)
		ti.Start = &val
	}

	return ti, true
}

func ExtractFromAndTo(input map[string]interface{}) (from, to AddressInfo, ok bool) {
	defer func() {
		if r := recover(); r != nil {
			ok = false // 防止 panic，可选
		}
	}()

	data, exists := input["data"].(map[string]interface{})
	if !exists {
		return AddressInfo{}, AddressInfo{}, false
	}

	fromMap, fromExists := data["from"].(map[string]interface{})
	toMap, toExists := data["to"].(map[string]interface{})

	if !fromExists || !toExists {
		return AddressInfo{}, AddressInfo{}, false
	}

	from = mapToAddressInfo(fromMap)
	to = mapToAddressInfo(toMap)
	return from, to, true
}

func mapToAddressInfo(m map[string]interface{}) AddressInfo {
	a := AddressInfo{}

	// 必填或已知存在的字段（根据你的数据）
	a.Address = m["address"].(string)
	a.Alias = m["alias"].(string)
	a.Code = m["code"].(string)
	a.Coin = m["coin"].(string)
	a.Name = m["name"].(string)
	a.Network = m["network"].(string)
	a.Tag = m["tag"].(string)

	// 可选字段（可能为 nil）
	if v := m["addressAlt"]; v != nil {
		a.AddressAlt = v
	}
	if amt, ok := m["amount"].(string); ok {
		a.Amount = &amt
	}
	if mc, ok := m["maxConfirmations"].(int); ok {
		a.MaxConfirmations = mc
	}
	if rc, ok := m["reqConfirmations"].(int); ok {
		a.ReqConfirmations = rc
	}
	a.TagName = m["tagName"] // 可能是 nil

	// Tx 字段（全 nil，但结构要对）
	txMap := m["tx"].(map[string]interface{})
	a.Tx = TxInfo{
		Amount:        txMap["amount"],
		CcyFee:        txMap["ccyfee"],
		Confirmations: txMap["confirmations"],
		Fee:           txMap["fee"],
		ID:            txMap["id"],
		TimeBlock:     txMap["timeBlock"],
		TimeReg:       txMap["timeReg"],
	}

	return a
}
func ExtractIDAndStatus(input map[string]interface{}) (id string, status string, ok bool) {
	// 安全访问 data 字段
	data, exists := input["data"].(map[string]interface{})
	if !exists {
		return "", "", false
	}

	// 提取 id 和 status
	idVal, idExists := data["id"].(string)
	statusVal, statusExists := data["status"].(string)

	if !idExists || !statusExists {
		return "", "", false
	}

	return idVal, statusVal, true
}
