package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type Client struct {
	db *sql.DB
}

func NewClient() (*Client, error) {
	db, err := sql.Open("sqlite3", "./unifra.db")
	if err != nil {
		return nil, err
	}

	client := &Client{db: db}
	// err = client.CreateAccountsTable()
	// if err != nil {
	// 	return nil, err
	// }
	return client, nil
}

func (c *Client) Close() error {
	return c.db.Close()
}

// func (c *Client) CreateAccountsTable() error {
// 	_, err := c.db.Exec(`
// 		CREATE TABLE IF NOT EXISTS accounts (
//             id INTEGER PRIMARY KEY AUTOINCREMENT,
//             address TEXT NOT NULL,
// 			private_key TEXT NOT NULL,
// 			nonce INTEGER NOT NULL
//         );
//     `)
// 	return err
// }

// func (c *Client) AddAccount(account types.Account) error {
// 	_, err := c.db.Exec("INSERT INTO accounts (address, private_key, nonce) VALUES (?, ?, ?)", account.Address, account.PrivateKey, account.Nonce)
// 	return err
// }

// func (c *Client) GetAccounts() ([]types.Account, error) {
// 	rows, err := c.db.Query("SELECT address, private_key, nonce FROM accounts ORDER BY address ASC")
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	var accounts []types.Account
// 	for rows.Next() {
// 		var account types.Account
// 		err := rows.Scan(&account.Address, &account.PrivateKey, &account.Nonce)
// 		if err != nil {
// 			return nil, err
// 		}
// 		accounts = append(accounts, account)
// 	}

// 	if err := rows.Err(); err != nil {
// 		return nil, err
// 	}

// 	return accounts, nil
// }
