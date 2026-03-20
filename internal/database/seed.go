package database

import (
	"database/sql"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const defaultSeedUserPassword = "Seed@123"

var firstNames = []string{"Ana", "Bruno", "Carla", "Diego", "Erika", "Felipe", "Gabriela", "Henrique", "Isabela", "Joao", "Karen", "Lucas", "Marina", "Nicolas", "Olivia", "Paulo", "Renata", "Samuel", "Talita", "Vinicius"}
var lastNames = []string{"Silva", "Souza", "Oliveira", "Costa", "Pereira", "Rodrigues", "Almeida", "Santos", "Lima", "Ferreira"}
var roles = []string{"employee", "manager", "admin"}
var departments = []string{"TI", "RH", "Financeiro", "Operacoes", "Comercial"}
var locations = []string{"Matriz", "Filial-SP", "Filial-RJ", "Remoto"}
var shifts = []string{"comercial", "manhã", "tarde", "noturno"}
var userTypes = []string{"internal", "external"}

func SeedUsers(db *sql.DB, count int) (int, error) {
	if count <= 0 {
		return 0, fmt.Errorf("a quantidade de usuários deve ser maior que zero")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(defaultSeedUserPassword), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("erro ao gerar hash da senha do usuário seed: %w", err)
	}

	query := `
		INSERT INTO users (
			name,
			type,
			email,
			password,
			role,
			department,
			phone,
			location,
			shift,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		ON CONFLICT (email) DO NOTHING`

	tx, err := db.Begin()
	if err != nil {
		return 0, fmt.Errorf("erro ao iniciar transação do seed de usuários: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	stmt, err := tx.Prepare(query)
	if err != nil {
		return 0, fmt.Errorf("erro ao preparar statement do seed de usuários: %w", err)
	}
	defer stmt.Close()

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	baseTimestamp := time.Now().UnixNano()
	inserted := 0

	for i := 0; i < count; i++ {
		firstName := randomItem(firstNames, rng)
		lastName := randomItem(lastNames, rng)
		name := firstName + " " + lastName
		email := randomEmail(firstName, lastName, baseTimestamp, i, rng)
		role := randomItem(roles, rng)
		department := randomItem(departments, rng)
		location := randomItem(locations, rng)
		shift := randomItem(shifts, rng)
		userType := randomItem(userTypes, rng)
		phone := fmt.Sprintf("11%08d", rng.Intn(100000000))

		result, execErr := stmt.Exec(
			name,
			userType,
			email,
			string(hashedPassword),
			role,
			department,
			phone,
			location,
			shift,
		)
		if execErr != nil {
			return 0, fmt.Errorf("erro ao executar seed de usuários: %w", execErr)
		}

		rowsAffected, rowsErr := result.RowsAffected()
		if rowsErr != nil {
			return 0, fmt.Errorf("erro ao verificar resultado do seed de usuários: %w", rowsErr)
		}

		inserted += int(rowsAffected)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("erro ao commitar seed de usuários: %w", err)
	}

	return inserted, nil
}

func randomItem(items []string, rng *rand.Rand) string {
	return items[rng.Intn(len(items))]
}

func randomEmail(firstName, lastName string, baseTimestamp int64, index int, rng *rand.Rand) string {
	base := strings.ToLower(firstName + "." + lastName)
	base = strings.ReplaceAll(base, " ", "")
	randomSuffix := rng.Intn(100000)
	return fmt.Sprintf("%s.%d.%d.%05d@timecontrol.local", base, baseTimestamp, index, randomSuffix)
}
