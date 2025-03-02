package fancylog

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

const (
	ColorReset         = "\033[0m"                // Сброс цвета
	ColorBlack         = "\033[30m"               // Чёрный
	ColorRed           = "\033[31m"               // Красный
	ColorGreen         = "\033[32m"               // Зелёный
	ColorYellow        = "\033[33m"               // Жёлтый
	ColorBlue          = "\033[34m"               // Синий
	ColorMagenta       = "\033[35m"               // Фиолетовый
	ColorCyan          = "\033[36m"               // Голубой
	ColorWhite         = "\033[37m"               // Белый
	ColorBrightBlack   = "\033[38;5;0m"           // Яркий чёрный
	ColorBrightRed     = "\033[38;5;9m"           // Яркий красный
	ColorBrightGreen   = "\033[38;5;10m"          // Яркий зелёный
	ColorBrightYellow  = "\033[38;5;11m"          // Яркий жёлтый
	ColorBrightBlue    = "\033[38;5;12m"          // Яркий синий
	ColorBrightMagenta = "\033[38;5;13m"          // Яркий фиолетовый
	ColorBrightCyan    = "\033[38;5;14m"          // Яркий голубой
	ColorBrightWhite   = "\033[38;5;15m"          // Яркий белый
	ColorOrange        = "\033[38;5;208m"         // Оранжевый
	ColorPink          = "\033[38;5;206m"         // Розовый
	ColorLime          = "\033[38;5;118m"         // Лайм
	ColorTeal          = "\033[38;5;30m"          // Бирюзовый
	ColorPurple        = "\033[38;5;93m"          // Пурпурный
	ColorMaroon        = "\033[38;5;88m"          // Бордовый
	ColorGray          = "\033[38;5;242m"         // Серый
	ColorLightGray     = "\033[38;5;250m"         // Светло-серый
	ColorCustomRed     = "\033[38;2;255;0;0m"     // Ярко-красный
	ColorCustomGreen   = "\033[38;2;0;255;0m"     // Ярко-зелёный
	ColorCustomBlue    = "\033[38;2;0;0;255m"     // Ярко-синий
	ColorCustomPink    = "\033[38;2;255;105;180m" // Розовый
	ColorCustomGold    = "\033[38;2;255;215;0m"   // Золотой
	ColorCustomSky     = "\033[38;2;135;206;250m" // Небесно-голубой
)

func Format(r io.Reader, w io.Writer, highlightConfig Config) {
	decoder := json.NewDecoder(r)
	for decoder.More() {
		var logEntry map[string]interface{}
		if err := decoder.Decode(&logEntry); err != nil {
			fmt.Fprintf(os.Stderr, "Error decoding JSON: %v\n", err)
			return
		}

		// Определяем максимальную длину ключа
		allKeys := make([]string, 0, len(logEntry))
		for key := range logEntry {
			allKeys = append(allKeys, key)
		}
		maxKeyLength := calculateMaxKeyLength(allKeys)

		// Создаём StringBuilder для форматированного вывода
		sb := new(strings.Builder)
		sb.WriteString("\n{\n")

		// Вызываем addFields с FieldOrder из highlightConfig
		addFields(sb, logEntry, maxKeyLength, highlightConfig)

		// Закрываем объект JSON
		sb.WriteString("\n}")
		fmt.Fprintln(w, sb.String())
	}
}

// calculateMaxKeyLength вычисляет максимальную длину ключа.
func calculateMaxKeyLength(keys []string) int {
	max := 0
	for _, key := range keys {
		if len(key)+2 > max { // +2 для учета кавычек
			max = len(key) + 2
		}
	}
	return max
}

func addFields(sb *strings.Builder, logEntry map[string]interface{}, maxKeyLength int, config Config) {
	// Создаём множество для отслеживания уже обработанных ключей
	processedKeys := make(map[string]bool)

	// Шаг 1: Выводим поля в заданном порядке (FieldOrder)
	fieldCount := len(config.FieldOrder)
	for i, key := range config.FieldOrder {
		if value, exists := logEntry[key]; exists {
			addField(sb, key, value, maxKeyLength, config.Rules[key])
			processedKeys[key] = true

			// Добавляем \n после каждого поля, кроме последнего
			if i < fieldCount-1 || len(logEntry) > fieldCount {
				sb.WriteString("\n")
			}
		}
	}

	// Шаг 2: Добавляем остальные поля в конце
	var remainingKeys []string
	for key := range logEntry {
		if !processedKeys[key] {
			remainingKeys = append(remainingKeys, key)
		}
	}

	// Сортируем оставшиеся ключи для предсказуемого порядка (опционально)
	sort.Strings(remainingKeys)

	// Выводим оставшиеся поля
	remainingCount := len(remainingKeys)
	for i, key := range remainingKeys {
		addField(sb, key, logEntry[key], maxKeyLength, config.Rules[key])

		// Добавляем \n после каждого поля, кроме последнего
		if i < remainingCount-1 {
			sb.WriteString("\n")
		}
	}
}

func addField(sb *strings.Builder, key string, value interface{}, maxKeyLength int, rule ColorRule) {
	paddedKey := fmt.Sprintf("%s%*s", key, maxKeyLength-len(key)-2, "")
	sb.WriteString("  ")
	sb.WriteString(paddedKey)
	sb.WriteString(": ")

	// Определяем цвет
	color := getColor(key, value, rule)

	switch v := value.(type) {
	case string:
		lines := wrapText(v, 80)
		for i, line := range lines {
			if color != "" {
				sb.WriteString(fmt.Sprintf("%s%s%s", color, line, ColorReset))
			} else {
				sb.WriteString(line)
			}
			if i < len(lines)-1 {
				sb.WriteString("\n" + strings.Repeat(" ", maxKeyLength+2))
			}
		}
	case float64, int, int64:
		if color != "" {
			sb.WriteString(fmt.Sprintf("%s%v%s", color, v, ColorReset))
		} else {
			sb.WriteString(fmt.Sprintf("%v", v))
		}
	default:
		if color != "" {
			sb.WriteString(fmt.Sprintf("%s%v%s", color, value, ColorReset))
		} else {
			sb.WriteString(fmt.Sprint(value))
		}
	}
}

type Config struct {
	FieldOrder []string   // Порядок вывода полей
	Rules      ColorsRule // Правила подсветки для каждого поля
}

type ColorsRule map[string]ColorRule

type ColorRule struct {
	DefaultColor string // Цвет по умолчанию для этого поля
	ValueColors  Colors // Специальные цвета для конкретных значений
}

type Colors map[string]string

func getColor(key string, value interface{}, rule ColorRule) string {
	if rule.DefaultColor == "" {
		return ""
	}

	strValue := fmt.Sprint(value)

	if rule.ValueColors != nil {
		if color, ok := rule.ValueColors[strValue]; ok {
			return color
		}
	}

	return rule.DefaultColor
}

// contains проверяет, содержит ли слайс определенное значение.
func contains(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

// wrapText разбивает строку на строки заданной длины.
func wrapText(text string, maxLen int) []string {
	var result []string
	words := strings.Fields(text)
	currentLine := ""

	for _, word := range words {
		if len(word) > maxLen {
			if currentLine != "" {
				result = append(result, currentLine)
				currentLine = ""
			}
			result = append(result, word)
			continue
		}

		if len(currentLine)+len(word)+1 > maxLen {
			result = append(result, currentLine)
			currentLine = word
		} else {
			if currentLine != "" {
				currentLine += " "
			}
			currentLine += word
		}
	}

	if currentLine != "" {
		result = append(result, currentLine)
	}

	return result
}
