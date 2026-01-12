package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	UUID       string    `json:"uuid" gorm:"primaryKey;type:varchar(36)"`
	Username   string    `json:"username" gorm:"not null"`
	TelegramID int64     `json:"telegram_id" gorm:"unique"`
	Balance    int       `json:"balance" gorm:"default:0"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Table struct {
	ID        int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Category  string    `json:"category" gorm:"type:varchar(10);check:category IN ('LOW','MID','VIP')"`
	Blinds    string    `json:"blinds" gorm:"type:varchar(20);not null"`
	BuyIn     int       `json:"buy_in" gorm:"not null"`
	Players   int       `json:"players" gorm:"default:0"`
	MaxSeats  int       `json:"max_seats" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TablePlayer struct {
	ID         int       `json:"id" gorm:"primaryKey;autoIncrement"`
	TableID    int       `json:"table_id" gorm:"not null"`
	UserUUID   string    `json:"user_uuid" gorm:"type:varchar(36);not null"`
	SeatNumber int       `json:"seat_number"`
	Chips      int       `json:"chips" gorm:"default:0"`
	JoinedAt   time.Time `json:"joined_at"`
	
	// Связи
	Table Table `json:"table" gorm:"foreignKey:TableID"`
	User  User  `json:"user" gorm:"foreignKey:UserUUID;references:UUID"`
}

type TableResponse struct {
	Tables []Table `json:"tables"`
}

// BeforeCreate хук для установки времени создания
func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	return nil
}

func (t *Table) BeforeCreate(tx *gorm.DB) error {
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()
	return nil
}

func (tp *TablePlayer) BeforeCreate(tx *gorm.DB) error {
	tp.JoinedAt = time.Now()
	return nil
}

// BeforeUpdate хук для обновления времени изменения
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	u.UpdatedAt = time.Now()
	return nil
}

func (t *Table) BeforeUpdate(tx *gorm.DB) error {
	t.UpdatedAt = time.Now()
	return nil
}
// Игровые модели для покера

type GameState string

const (
	GameStateWaiting    GameState = "waiting"
	GameStatePreFlop    GameState = "preflop"
	GameStateFlop       GameState = "flop"
	GameStateTurn       GameState = "turn"
	GameStateRiver      GameState = "river"
	GameStateShowdown   GameState = "showdown"
	GameStateFinished   GameState = "finished"
)

type PlayerAction string

const (
	ActionFold  PlayerAction = "fold"
	ActionCall  PlayerAction = "call"
	ActionRaise PlayerAction = "raise"
	ActionCheck PlayerAction = "check"
	ActionBet   PlayerAction = "bet"
)

type Card struct {
	Suit  string `json:"suit"`  // hearts, diamonds, clubs, spades
	Rank  string `json:"rank"`  // 2-10, J, Q, K, A
	Value int    `json:"value"` // 2-14 (A=14)
}

type Game struct {
	ID            string      `json:"id" gorm:"primaryKey;type:varchar(36)"`
	TableID       int         `json:"table_id" gorm:"not null"`
	State         GameState   `json:"state" gorm:"type:varchar(20);default:'waiting'"`
	Deck          []Card      `json:"deck" gorm:"type:jsonb"`
	CommunityCards []Card     `json:"community_cards" gorm:"type:jsonb"`
	Pot           int         `json:"pot" gorm:"default:0"`
	CurrentBet    int         `json:"current_bet" gorm:"default:0"`
	DealerPosition int        `json:"dealer_position" gorm:"default:0"`
	CurrentPlayer int         `json:"current_player" gorm:"default:0"`
	SmallBlind    int         `json:"small_blind"`
	BigBlind      int         `json:"big_blind"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
	
	// Связи
	Table   Table         `json:"table" gorm:"foreignKey:TableID"`
	Players []GamePlayer  `json:"players" gorm:"foreignKey:GameID"`
}

type GamePlayer struct {
	ID         int          `json:"id" gorm:"primaryKey;autoIncrement"`
	GameID     string       `json:"game_id" gorm:"type:varchar(36);not null"`
	UserUUID   string       `json:"user_uuid" gorm:"type:varchar(36);not null"`
	Position   int          `json:"position" gorm:"not null"`
	Cards      []Card       `json:"cards" gorm:"type:jsonb"`
	Chips      int          `json:"chips" gorm:"default:0"`
	Bet        int          `json:"bet" gorm:"default:0"`
	IsFolded   bool         `json:"is_folded" gorm:"default:false"`
	IsAllIn    bool         `json:"is_all_in" gorm:"default:false"`
	LastAction PlayerAction `json:"last_action" gorm:"type:varchar(10)"`
	
	// Связи
	Game Game `json:"game" gorm:"foreignKey:GameID"`
	User User `json:"user" gorm:"foreignKey:UserUUID;references:UUID"`
}

type GameAction struct {
	ID        int          `json:"id" gorm:"primaryKey;autoIncrement"`
	GameID    string       `json:"game_id" gorm:"type:varchar(36);not null"`
	UserUUID  string       `json:"user_uuid" gorm:"type:varchar(36);not null"`
	Action    PlayerAction `json:"action" gorm:"type:varchar(10);not null"`
	Amount    int          `json:"amount" gorm:"default:0"`
	CreatedAt time.Time    `json:"created_at"`
	
	// Связи
	Game Game `json:"game" gorm:"foreignKey:GameID"`
	User User `json:"user" gorm:"foreignKey:UserUUID;references:UUID"`
}

// Kafka сообщения
type GameEvent struct {
	Type      string      `json:"type"`
	GameID    string      `json:"game_id"`
	TableID   int         `json:"table_id"`
	UserUUID  string      `json:"user_uuid,omitempty"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

type PlayerActionEvent struct {
	Action PlayerAction `json:"action"`
	Amount int          `json:"amount"`
	Player GamePlayer   `json:"player"`
}

type GameStateEvent struct {
	State          GameState `json:"state"`
	CommunityCards []Card    `json:"community_cards"`
	Pot            int       `json:"pot"`
	CurrentBet     int       `json:"current_bet"`
	CurrentPlayer  int       `json:"current_player"`
}

// Хуки для Game
func (g *Game) BeforeCreate(tx *gorm.DB) error {
	if g.ID == "" {
		g.ID = uuid.New().String()
	}
	g.CreatedAt = time.Now()
	g.UpdatedAt = time.Now()
	return nil
}

func (g *Game) BeforeUpdate(tx *gorm.DB) error {
	g.UpdatedAt = time.Now()
	return nil
}

func (ga *GameAction) BeforeCreate(tx *gorm.DB) error {
	ga.CreatedAt = time.Now()
	return nil
}