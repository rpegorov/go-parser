package services

type EnterpriseService interface {
	ParseAndSaveEnterpriseTree(body []byte) (Result, error)
}

type Result struct {
	Enterprises int
	Sites       int
	Departments int
	Equipment   int
}
