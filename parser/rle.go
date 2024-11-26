package parser

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Rule struct {
	birth    map[uint]struct{}
	survival map[uint]struct{}
}

func (r *Rule) IsBirth(n uint) bool {
	_, exist := r.birth[n]
	return exist
}

func (r *Rule) IsSurvival(n uint) bool {
	_, exist := r.survival[n]
	return exist
}

func ParseRule(str string) (*Rule, error) {
	rule := &Rule{
		birth:    make(map[uint]struct{}),
		survival: make(map[uint]struct{}),
	}

	parts := strings.Split(str, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("無効なルール形式: %s", str)
	}

	for _, part := range parts {
		if strings.HasPrefix(part, "B") {
			for _, ch := range part[1:] {
				num, err := strconv.Atoi(string(ch))
				if err != nil {
					return nil, fmt.Errorf("無効な誕生条件: %s", string(ch))
				}
				rule.birth[uint(num)] = struct{}{}
			}
		} else if strings.HasPrefix(part, "S") {
			for _, ch := range part[1:] {
				num, err := strconv.Atoi(string(ch))
				if err != nil {
					return nil, fmt.Errorf("無効な生存条件: %s", string(ch))
				}
				rule.survival[uint(num)] = struct{}{}
			}
		} else {
			return nil, fmt.Errorf("無効なルール形式: %s", part)
		}
	}

	return rule, nil
}

func LoadRLE(filename string, size int) ([][]uint, *Rule, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	header := ""
	data := ""
	var rule *Rule

	for scanner.Scan() {
		line := scanner.Text()

		// #で始まるコメント行はスキップする
		if strings.HasPrefix(line, "#") {
			continue
		}

		if header == "" {
			header = line
			re := regexp.MustCompile(`x\s*=\s*(\d+),\s*y\s*=\s*(\d+)(?:,\s*rule\s*=\s*(\S+))?`)
			matches := re.FindStringSubmatch(line)
			if matches == nil {
				return nil, nil, fmt.Errorf("無効なヘッダー形式: %s", line)
			}
			rule, err = ParseRule(matches[3])
			if err != nil {
				return nil, nil, err
			}
			continue
		}

		data += line
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}

	board, err := parseRLE(data, size)
	if err != nil {
		return nil, nil, err
	}

	return board, rule, nil
}

func parseRLE(data string, size int) ([][]uint, error) {
	world := make([][]uint, size)
	for i := 0; i < size; i++ {
		world[i] = make([]uint, size)
	}
	x := 0
	y := 0
	num := ""

	for _, ch := range data {
		switch {
		case ch >= '0' && ch <= '9':
			num += string(ch)
		case ch == 'b':
			count := 1
			if num != "" {
				var err error
				count, err = strconv.Atoi(num)
				if err != nil {
					return nil, fmt.Errorf("無効な数値: %s", num)
				}
				num = ""
			}
			for i := 0; i < count; i++ {
				if x >= size || y >= size {
					return nil, fmt.Errorf("パターンがサイズを超えています")
				}
				world[x][y] = 0
				x++
			}
		case ch == 'o':
			count := 1
			if num != "" {
				var err error
				count, err = strconv.Atoi(num)
				if err != nil {
					return nil, fmt.Errorf("無効な数値: %s", num)
				}
				num = ""
			}
			for i := 0; i < count; i++ {
				if x >= size || y >= size {
					return nil, fmt.Errorf("パターンがサイズを超えています")
				}
				world[x][y] = 1
				x++
			}
		case ch == '$':
			count := 1
			if num != "" {
				var err error
				count, err = strconv.Atoi(num)
				if err != nil {
					return nil, fmt.Errorf("無効な数値: %s", num)
				}
				num = ""
			}
			y += count
			x = 0
		case ch == '!':
			return world, nil
		}
	}

	return world, nil
}