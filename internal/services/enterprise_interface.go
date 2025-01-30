package services

type EnterpriseService interface {
	ParseAndSaveEnterpriseTree(body []byte) (EnterpriseTree, error)
}

type EnterpriseTree struct {
	Enterprises int
	Sites       int
	Departments int
	Equipment   int
}
