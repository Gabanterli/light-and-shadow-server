package economy

import (
	"errors"
	"fmt"
)

// Constantes de conversão de moedas baseadas no Bronze como unidade fundamental (1 Bronze)
const (
	BronzePerSilver  int64 = 1000
	BronzePerGold    int64 = 1000 * 1000
	BronzePerDiamond int64 = 1000 * 1000 * 1000
)

// Tiers representa o valor quebrado nas 4 denominações de moedas
type Tiers struct {
	Diamond int64 `json:"diamond"`
	Gold    int64 `json:"gold"`
	Silver  int64 `json:"silver"`
	Bronze  int64 `json:"bronze"`
}

// ConvertToTiers converte o saldo total armazenado em Bronze (int64) para os 4 tiers de forma determinística
func ConvertToTiers(totalBronze int64) Tiers {
	if totalBronze < 0 {
		totalBronze = 0
	}
	diamond := totalBronze / BronzePerDiamond
	rem := totalBronze % BronzePerDiamond

	gold := rem / BronzePerGold
	rem = rem % BronzePerGold

	silver := rem / BronzePerSilver
	bronze := rem % BronzePerSilver

	return Tiers{
		Diamond: diamond,
		Gold:    gold,
		Silver:  silver,
		Bronze:  bronze,
	}
}

// ConvertFromTiers consolida as moedas denominadas em Diamond, Gold, Silver e Bronze de volta ao valor de Bronze int64 único
func ConvertFromTiers(diamond, gold, silver, bronze int64) (int64, error) {
	if diamond < 0 || gold < 0 || silver < 0 || bronze < 0 {
		return 0, errors.New("currency values cannot be negative")
	}
	return (diamond * BronzePerDiamond) + (gold * BronzePerGold) + (silver * BronzePerSilver) + bronze, nil
}

// ValidateBalance garante que um saldo de moeda não seja negativo
func ValidateBalance(balance int64) error {
	if balance < 0 {
		return errors.New("invalid balance: currency cannot be negative")
	}
	return nil
}

// FormatCurrency formata o saldo total de forma limpa para exibição textual (ex: "5d 230g 40s 950b")
func FormatCurrency(totalBronze int64) string {
	if totalBronze < 0 {
		return "0b"
	}
	tiers := ConvertToTiers(totalBronze)
	if tiers.Diamond > 0 {
		return fmt.Sprintf("%dd %dg %ds %db", tiers.Diamond, tiers.Gold, tiers.Silver, tiers.Bronze)
	}
	if tiers.Gold > 0 {
		return fmt.Sprintf("%dg %ds %db", tiers.Gold, tiers.Silver, tiers.Bronze)
	}
	if tiers.Silver > 0 {
		return fmt.Sprintf("%ds %db", tiers.Silver, tiers.Bronze)
	}
	return fmt.Sprintf("%db", tiers.Bronze)
}

// LegacyGoldToBronze auxilia na conversão de valores legados de ouro (Gold) para o equivalente em Bronze
func LegacyGoldToBronze(legacyGold int) int64 {
	if legacyGold < 0 {
		return 0
	}
	return int64(legacyGold) * BronzePerGold
}

// BronzeToLegacyGold converte o saldo de bronze de volta para uma aproximação de Gold legada (truncada para compatibilidade)
func BronzeToLegacyGold(bronze int64) int {
	if bronze < 0 {
		return 0
	}
	legacyGold := bronze / BronzePerGold
	return int(legacyGold)
}
