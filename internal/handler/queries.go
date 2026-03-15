package handler

var (
	GetCompaniesKey   = "GetCompanies"
	GetCompaniesQuery = `
		SELECT id, name, description, employees, registered, company_type, created_at, updated_at
		FROM companies;
	`

	GetCompanyKey   = "GetCompany"
	GetCompanyQuery = `
		SELECT id, name, description, employees, registered, company_type, created_at, updated_at
		FROM companies
		WHERE id = $1;
	`

	CreateCompanyKey   = "CreateCompany"
	CreateCompanyQuery = `
		INSERT INTO companies (id, name, description, employees, registered, company_type)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, name, description, employees, registered, company_type, created_at, updated_at;
	`

	UpdateCompanyKey   = "UpdateCompany"
	UpdateCompanyQuery = `
		UPDATE companies
		SET 
			name = COALESCE($2, name), 
			description = COALESCE($3, description), 
			employees = COALESCE($4, employees), 
			registered = COALESCE($5, registered), 
			company_type = COALESCE($6, company_type), 
			updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, description, employees, registered, company_type, created_at, updated_at;
	`

	DeleteCompanyKey   = "DeleteCompany"
	DeleteCompanyQuery = `
		DELETE FROM companies
		WHERE id = $1;
	`

	LoginKey      = "Login"
	LoginKeyQuery = `
		SELECT id
		FROM users
		WHERE username = $1 AND password = $2;
	`
)
