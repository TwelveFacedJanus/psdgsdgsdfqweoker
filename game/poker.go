package game

import (
	"fmt"
	"math/rand"
	"sort"
	"time"

	"poker/models"
)

// PokerEngine основной движок игры в покер
type PokerEngine struct {
	game *models.Game
}

// NewPokerEngine создает новый движок игры
func NewPokerEngine(game *models.Game) *PokerEngine {
	return &PokerEngine{game: game}
}

// CreateDeck создает новую колоду карт
func CreateDeck() []models.Card {
	suits := []string{"hearts", "diamonds", "clubs", "spades"}
	ranks := []string{"2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K", "A"}
	values := []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14}

	var deck []models.Card
	for _, suit := range suits {
		for i, rank := range ranks {
			deck = append(deck, models.Card{
				Suit:  suit,
				Rank:  rank,
				Value: values[i],
			})
		}
	}

	// Перемешиваем колоду
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})

	return deck
}

// DealCards раздает карты игрокам
func (pe *PokerEngine) DealCards() {
	// Каждому игроку по 2 карты
	for i := range pe.game.Players {
		if !pe.game.Players[i].IsFolded {
			pe.game.Players[i].Cards = []models.Card{
				pe.game.Deck[0],
				pe.game.Deck[1],
			}
			pe.game.Deck = pe.game.Deck[2:]
		}
	}
}

// DealFlop раздает флоп (3 карты)
func (pe *PokerEngine) DealFlop() {
	// Сжигаем одну карту
	pe.game.Deck = pe.game.Deck[1:]
	
	// Добавляем 3 карты к общим
	pe.game.CommunityCards = append(pe.game.CommunityCards, 
		pe.game.Deck[0], pe.game.Deck[1], pe.game.Deck[2])
	pe.game.Deck = pe.game.Deck[3:]
}

// DealTurn раздает терн (1 карта)
func (pe *PokerEngine) DealTurn() {
	// Сжигаем одну карту
	pe.game.Deck = pe.game.Deck[1:]
	
	// Добавляем 1 карту к общим
	pe.game.CommunityCards = append(pe.game.CommunityCards, pe.game.Deck[0])
	pe.game.Deck = pe.game.Deck[1:]
}

// DealRiver раздает ривер (1 карта)
func (pe *PokerEngine) DealRiver() {
	// Сжигаем одну карту
	pe.game.Deck = pe.game.Deck[1:]
	
	// Добавляем 1 карту к общим
	pe.game.CommunityCards = append(pe.game.CommunityCards, pe.game.Deck[0])
	pe.game.Deck = pe.game.Deck[1:]
}

// GetNextPlayer возвращает следующего активного игрока
func (pe *PokerEngine) GetNextPlayer() int {
	activePlayers := pe.GetActivePlayers()
	if len(activePlayers) <= 1 {
		return -1
	}

	currentIndex := -1
	for i, player := range activePlayers {
		if player.Position == pe.game.CurrentPlayer {
			currentIndex = i
			break
		}
	}

	if currentIndex == -1 {
		return activePlayers[0].Position
	}

	nextIndex := (currentIndex + 1) % len(activePlayers)
	return activePlayers[nextIndex].Position
}

// GetActivePlayers возвращает активных игроков (не сфолдивших)
func (pe *PokerEngine) GetActivePlayers() []models.GamePlayer {
	var active []models.GamePlayer
	for _, player := range pe.game.Players {
		if !player.IsFolded {
			active = append(active, player)
		}
	}
	return active
}

// CanPlayerAct проверяет, может ли игрок действовать
func (pe *PokerEngine) CanPlayerAct(userUUID string) bool {
	for _, player := range pe.game.Players {
		if player.UserUUID == userUUID && player.Position == pe.game.CurrentPlayer {
			return !player.IsFolded && !player.IsAllIn
		}
	}
	return false
}

// ProcessAction обрабатывает действие игрока
func (pe *PokerEngine) ProcessAction(userUUID string, action models.PlayerAction, amount int) error {
	if !pe.CanPlayerAct(userUUID) {
		return fmt.Errorf("игрок не может действовать")
	}

	playerIndex := -1
	for i, player := range pe.game.Players {
		if player.UserUUID == userUUID {
			playerIndex = i
			break
		}
	}

	if playerIndex == -1 {
		return fmt.Errorf("игрок не найден")
	}

	player := &pe.game.Players[playerIndex]

	switch action {
	case models.ActionFold:
		player.IsFolded = true
		player.LastAction = models.ActionFold

	case models.ActionCall:
		callAmount := pe.game.CurrentBet - player.Bet
		if callAmount > player.Chips {
			callAmount = player.Chips
			player.IsAllIn = true
		}
		player.Chips -= callAmount
		player.Bet += callAmount
		pe.game.Pot += callAmount
		player.LastAction = models.ActionCall

	case models.ActionRaise:
		if amount < pe.game.CurrentBet*2 {
			return fmt.Errorf("размер рейза слишком мал")
		}
		raiseAmount := amount - player.Bet
		if raiseAmount > player.Chips {
			raiseAmount = player.Chips
			player.IsAllIn = true
		}
		player.Chips -= raiseAmount
		player.Bet += raiseAmount
		pe.game.Pot += raiseAmount
		pe.game.CurrentBet = player.Bet
		player.LastAction = models.ActionRaise

	case models.ActionCheck:
		if pe.game.CurrentBet > player.Bet {
			return fmt.Errorf("нельзя чекать, есть ставка")
		}
		player.LastAction = models.ActionCheck

	case models.ActionBet:
		if pe.game.CurrentBet > 0 {
			return fmt.Errorf("нельзя ставить, уже есть ставка")
		}
		if amount > player.Chips {
			amount = player.Chips
			player.IsAllIn = true
		}
		player.Chips -= amount
		player.Bet += amount
		pe.game.Pot += amount
		pe.game.CurrentBet = amount
		player.LastAction = models.ActionBet
	}

	// Переходим к следующему игроку
	pe.game.CurrentPlayer = pe.GetNextPlayer()

	return nil
}

// IsRoundComplete проверяет, завершен ли раунд торговли
func (pe *PokerEngine) IsRoundComplete() bool {
	activePlayers := pe.GetActivePlayers()
	if len(activePlayers) <= 1 {
		return true
	}

	// Проверяем, что все активные игроки сделали одинаковые ставки
	for _, player := range activePlayers {
		if !player.IsAllIn && player.Bet != pe.game.CurrentBet {
			return false
		}
	}

	return true
}

// AdvanceGameState переводит игру в следующее состояние
func (pe *PokerEngine) AdvanceGameState() {
	switch pe.game.State {
	case models.GameStateWaiting:
		pe.game.State = models.GameStatePreFlop
		pe.DealCards()
		
	case models.GameStatePreFlop:
		pe.game.State = models.GameStateFlop
		pe.DealFlop()
		pe.ResetBets()
		
	case models.GameStateFlop:
		pe.game.State = models.GameStateTurn
		pe.DealTurn()
		pe.ResetBets()
		
	case models.GameStateTurn:
		pe.game.State = models.GameStateRiver
		pe.DealRiver()
		pe.ResetBets()
		
	case models.GameStateRiver:
		pe.game.State = models.GameStateShowdown
		pe.DetermineWinner()
		
	case models.GameStateShowdown:
		pe.game.State = models.GameStateFinished
	}
}

// ResetBets сбрасывает ставки для нового раунда
func (pe *PokerEngine) ResetBets() {
	pe.game.CurrentBet = 0
	for i := range pe.game.Players {
		pe.game.Players[i].Bet = 0
		pe.game.Players[i].LastAction = ""
	}
	
	// Начинаем с игрока после дилера
	activePlayers := pe.GetActivePlayers()
	if len(activePlayers) > 0 {
		pe.game.CurrentPlayer = activePlayers[0].Position
	}
}

// DetermineWinner определяет победителя
func (pe *PokerEngine) DetermineWinner() {
	activePlayers := pe.GetActivePlayers()
	if len(activePlayers) == 1 {
		// Только один игрок остался
		winner := &activePlayers[0]
		winner.Chips += pe.game.Pot
		pe.game.Pot = 0
		return
	}

	// Определяем лучшие комбинации
	bestHands := make(map[string]HandRank)
	for _, player := range activePlayers {
		allCards := append(player.Cards, pe.game.CommunityCards...)
		bestHands[player.UserUUID] = GetBestHand(allCards)
	}

	// Находим победителя
	var winners []string
	var bestRank HandRank
	
	for userUUID, hand := range bestHands {
		if len(winners) == 0 || hand.Rank > bestRank.Rank {
			winners = []string{userUUID}
			bestRank = hand
		} else if hand.Rank == bestRank.Rank {
			// Сравниваем кикеры
			if compareKickers(hand.Kickers, bestRank.Kickers) > 0 {
				winners = []string{userUUID}
				bestRank = hand
			} else if compareKickers(hand.Kickers, bestRank.Kickers) == 0 {
				winners = append(winners, userUUID)
			}
		}
	}

	// Распределяем банк между победителями
	winAmount := pe.game.Pot / len(winners)
	for _, winnerUUID := range winners {
		for i := range pe.game.Players {
			if pe.game.Players[i].UserUUID == winnerUUID {
				pe.game.Players[i].Chips += winAmount
				break
			}
		}
	}
	
	pe.game.Pot = 0
}

// HandRank представляет ранг руки
type HandRank struct {
	Rank    int   `json:"rank"`    // 1-10 (1=старшая карта, 10=роял флеш)
	Kickers []int `json:"kickers"` // Кикеры для сравнения
}

// GetBestHand определяет лучшую комбинацию из 7 карт
func GetBestHand(cards []models.Card) HandRank {
	if len(cards) < 5 {
		return HandRank{Rank: 1, Kickers: []int{}}
	}

	// Сортируем карты по убыванию
	sort.Slice(cards, func(i, j int) bool {
		return cards[i].Value > cards[j].Value
	})

	// Проверяем все возможные комбинации
	if isRoyalFlush(cards) {
		return HandRank{Rank: 10, Kickers: []int{14}}
	}
	if rank := isStraightFlush(cards); rank.Rank > 0 {
		return rank
	}
	if rank := isFourOfAKind(cards); rank.Rank > 0 {
		return rank
	}
	if rank := isFullHouse(cards); rank.Rank > 0 {
		return rank
	}
	if rank := isFlush(cards); rank.Rank > 0 {
		return rank
	}
	if rank := isStraight(cards); rank.Rank > 0 {
		return rank
	}
	if rank := isThreeOfAKind(cards); rank.Rank > 0 {
		return rank
	}
	if rank := isTwoPair(cards); rank.Rank > 0 {
		return rank
	}
	if rank := isPair(cards); rank.Rank > 0 {
		return rank
	}

	// Старшая карта
	kickers := make([]int, 0, 5)
	for i := 0; i < 5 && i < len(cards); i++ {
		kickers = append(kickers, cards[i].Value)
	}
	return HandRank{Rank: 1, Kickers: kickers}
}

// Функции для определения комбинаций
func isRoyalFlush(cards []models.Card) bool {
	return isStraightFlush(cards).Rank > 0 && cards[0].Value == 14
}

func isStraightFlush(cards []models.Card) HandRank {
	flushSuit := getFlushSuit(cards)
	if flushSuit == "" {
		return HandRank{}
	}

	var flushCards []models.Card
	for _, card := range cards {
		if card.Suit == flushSuit {
			flushCards = append(flushCards, card)
		}
	}

	if straight := isStraight(flushCards); straight.Rank > 0 {
		return HandRank{Rank: 9, Kickers: straight.Kickers}
	}

	return HandRank{}
}

func isFourOfAKind(cards []models.Card) HandRank {
	counts := make(map[int]int)
	for _, card := range cards {
		counts[card.Value]++
	}

	var fourValue, kicker int
	for value, count := range counts {
		if count == 4 {
			fourValue = value
		} else if count >= 1 && value > kicker {
			kicker = value
		}
	}

	if fourValue > 0 {
		return HandRank{Rank: 8, Kickers: []int{fourValue, kicker}}
	}

	return HandRank{}
}

func isFullHouse(cards []models.Card) HandRank {
	counts := make(map[int]int)
	for _, card := range cards {
		counts[card.Value]++
	}

	var threeValue, pairValue int
	for value, count := range counts {
		if count >= 3 && value > threeValue {
			threeValue = value
		}
	}
	
	for value, count := range counts {
		if count >= 2 && value != threeValue && value > pairValue {
			pairValue = value
		}
	}

	if threeValue > 0 && pairValue > 0 {
		return HandRank{Rank: 7, Kickers: []int{threeValue, pairValue}}
	}

	return HandRank{}
}

func isFlush(cards []models.Card) HandRank {
	flushSuit := getFlushSuit(cards)
	if flushSuit == "" {
		return HandRank{}
	}

	var flushCards []models.Card
	for _, card := range cards {
		if card.Suit == flushSuit {
			flushCards = append(flushCards, card)
		}
	}

	sort.Slice(flushCards, func(i, j int) bool {
		return flushCards[i].Value > flushCards[j].Value
	})

	kickers := make([]int, 0, 5)
	for i := 0; i < 5 && i < len(flushCards); i++ {
		kickers = append(kickers, flushCards[i].Value)
	}

	return HandRank{Rank: 6, Kickers: kickers}
}

func isStraight(cards []models.Card) HandRank {
	// Убираем дубликаты
	uniqueValues := make(map[int]bool)
	for _, card := range cards {
		uniqueValues[card.Value] = true
	}

	var values []int
	for value := range uniqueValues {
		values = append(values, value)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(values)))

	// Проверяем стрит
	for i := 0; i <= len(values)-5; i++ {
		if values[i]-values[i+4] == 4 {
			return HandRank{Rank: 5, Kickers: []int{values[i]}}
		}
	}

	// Проверяем A-2-3-4-5 стрит
	if len(values) >= 5 && values[0] == 14 {
		lowStraight := []int{14, 5, 4, 3, 2}
		hasLowStraight := true
		for _, val := range lowStraight {
			if !uniqueValues[val] {
				hasLowStraight = false
				break
			}
		}
		if hasLowStraight {
			return HandRank{Rank: 5, Kickers: []int{5}}
		}
	}

	return HandRank{}
}

func isThreeOfAKind(cards []models.Card) HandRank {
	counts := make(map[int]int)
	for _, card := range cards {
		counts[card.Value]++
	}

	var threeValue int
	var kickers []int
	
	for value, count := range counts {
		if count == 3 {
			threeValue = value
		} else {
			for i := 0; i < count; i++ {
				kickers = append(kickers, value)
			}
		}
	}

	if threeValue > 0 {
		sort.Sort(sort.Reverse(sort.IntSlice(kickers)))
		if len(kickers) > 2 {
			kickers = kickers[:2]
		}
		return HandRank{Rank: 4, Kickers: append([]int{threeValue}, kickers...)}
	}

	return HandRank{}
}

func isTwoPair(cards []models.Card) HandRank {
	counts := make(map[int]int)
	for _, card := range cards {
		counts[card.Value]++
	}

	var pairs []int
	var kickers []int
	
	for value, count := range counts {
		if count >= 2 {
			pairs = append(pairs, value)
			if count > 2 {
				for i := 0; i < count-2; i++ {
					kickers = append(kickers, value)
				}
			}
		} else {
			for i := 0; i < count; i++ {
				kickers = append(kickers, value)
			}
		}
	}

	if len(pairs) >= 2 {
		sort.Sort(sort.Reverse(sort.IntSlice(pairs)))
		sort.Sort(sort.Reverse(sort.IntSlice(kickers)))
		
		if len(kickers) > 1 {
			kickers = kickers[:1]
		}
		
		return HandRank{Rank: 3, Kickers: append(pairs[:2], kickers...)}
	}

	return HandRank{}
}

func isPair(cards []models.Card) HandRank {
	counts := make(map[int]int)
	for _, card := range cards {
		counts[card.Value]++
	}

	var pairValue int
	var kickers []int
	
	for value, count := range counts {
		if count == 2 {
			pairValue = value
		} else {
			for i := 0; i < count; i++ {
				kickers = append(kickers, value)
			}
		}
	}

	if pairValue > 0 {
		sort.Sort(sort.Reverse(sort.IntSlice(kickers)))
		if len(kickers) > 3 {
			kickers = kickers[:3]
		}
		return HandRank{Rank: 2, Kickers: append([]int{pairValue}, kickers...)}
	}

	return HandRank{}
}

func getFlushSuit(cards []models.Card) string {
	suitCounts := make(map[string]int)
	for _, card := range cards {
		suitCounts[card.Suit]++
	}

	for suit, count := range suitCounts {
		if count >= 5 {
			return suit
		}
	}

	return ""
}

func compareKickers(kickers1, kickers2 []int) int {
	for i := 0; i < len(kickers1) && i < len(kickers2); i++ {
		if kickers1[i] > kickers2[i] {
			return 1
		} else if kickers1[i] < kickers2[i] {
			return -1
		}
	}
	return 0
}