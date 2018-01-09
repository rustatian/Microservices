package vault

type PasswrdBatch []string

type pswrd struct{}

func (p *pswrd) passwordHashProcess(batch PasswrdBatch) (hash []string, e error) {
	if len(batch) == 0 {
		return
	}

	//for i := 0; i <= len(batch); i++ {
	//	hash, err := bcrypt.GenerateFromPassword([]byte(batch[i]), bcrypt.DefaultCost)
	//	if err != nil {
	//		return "", err
	//	}
	//	return string(hash), nil
	//}

	return
}
