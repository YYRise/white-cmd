package parse

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/YYRise/white-cmd/exec"
)

var (
	ParseBacktick bool = false
)

func isSpace(r rune) bool {
	switch r {
	case ' ', '\t', '\r', '\n':
		return true
	}
	return false
}

type Parser struct {
	ParseBacktick bool
	Position      int
	Dir           string
}

func NewParser() *Parser {
	return &Parser{
		ParseBacktick: ParseBacktick,
		Position:      0,
		Dir:           "",
	}
}

// argParseState 定义解析器所处的状态
type argParseState int

const (
	StateNormal        argParseState = iota // 正常解析
	StateEscaped                            // 遇到转义字符 \
	StateSingleQuote                        // 在单引号内
	StateDoubleQuote                        // 在双引号内
	StateBacktick                           // 在反引号内
	StateDollarCommand                      // 在 $() 命令替换内
)

func (p *Parser) Parse(line string) ([]string, error) {
	args := []string{}
	var currentArgBuf bytes.Buffer  // 使用 bytes.Buffer 替代字符串拼接
	var backtickCmdBuf bytes.Buffer // 用于存储反引号或 $() 内的命令
	state := StateNormal

	pos := -1 // 记录特殊分隔符的位置

	runes := []rune(line)
	for i, r := range runes {
		switch state {
		case StateNormal:
			if isSpace(r) {
				if currentArgBuf.Len() > 0 { // 如果有内容，则添加为参数
					args = append(args, currentArgBuf.String())
					currentArgBuf.Reset() // 清空缓冲区
				}
				continue
			}

			switch r {
			case '\\':
				state = StateEscaped
			case '\'':
				state = StateSingleQuote
			case '"':
				state = StateDoubleQuote
			case '`':
				if p.ParseBacktick {
					state = StateBacktick
				} else {
					currentArgBuf.WriteRune(r) // 如果不支持反引号，视为普通字符
				}
			case '$': // 检查是否是 $()，否则视为普通字符
				if i+1 < len(runes) && runes[i+1] == '(' {
					if p.ParseBacktick {
						state = StateDollarCommand
						currentArgBuf.WriteRune(r)   // 写入 $
						currentArgBuf.WriteRune('(') // 写入 (
						i++                          // 跳过 (
					} else {
						currentArgBuf.WriteRune(r) // 如果不支持 $()，视为普通字符
					}
				} else {
					currentArgBuf.WriteRune(r) // 否则只是一个普通字符 $
				}
			case ';', '&', '|', '<', '>':
				// 遇到特殊分隔符，当前参数结束
				if currentArgBuf.Len() > 0 {
					args = append(args, currentArgBuf.String())
					currentArgBuf.Reset()
				}
				pos = i      // 记录分隔符位置
				goto endLoop // 结束循环
			default:
				currentArgBuf.WriteRune(r)
			}

		case StateEscaped:
			currentArgBuf.WriteRune(r)
			state = StateNormal

		case StateSingleQuote:
			if r == '\'' {
				state = StateNormal
			} else {
				currentArgBuf.WriteRune(r)
			}

		case StateDoubleQuote:
			if r == '"' {
				state = StateNormal
			} else if r == '\\' { // 双引号内也支持转义，但转义符会保留
				currentArgBuf.WriteRune(r)
				if i+1 < len(runes) { // 确保有下一个字符
					currentArgBuf.WriteRune(runes[i+1])
					i++ // 跳过转义的字符
				}
			} else {
				currentArgBuf.WriteRune(r)
			}

		case StateBacktick:
			if r == '`' {
				if p.ParseBacktick {
					cmd := backtickCmdBuf.String()
					out, err := exec.ExecSh(cmd, p.Dir)
					if err != nil {
						return nil, fmt.Errorf("backtick command execution failed: %w", err)
					}
					currentArgBuf.WriteString(out)
					backtickCmdBuf.Reset()
				} else {
					currentArgBuf.WriteRune('`') // 如果不支持，则保留反引号
					currentArgBuf.WriteString(backtickCmdBuf.String())
					currentArgBuf.WriteRune('`')
					backtickCmdBuf.Reset()
				}
				state = StateNormal
			} else {
				backtickCmdBuf.WriteRune(r)
			}

		case StateDollarCommand:
			if r == ')' {
				if p.ParseBacktick {
					cmd := backtickCmdBuf.String()
					out, err := exec.ExecSh(cmd, p.Dir)
					if err != nil {
						return nil, fmt.Errorf("dollar command execution failed: %w", err)
					}
					// 移除 currentArgBuf 中最后的 $ (
					currentArgBufStr := currentArgBuf.String()
					if strings.HasSuffix(currentArgBufStr, "$(") {
						currentArgBuf.Reset()
						currentArgBuf.WriteString(currentArgBufStr[:len(currentArgBufStr)-2])
					}
					currentArgBuf.WriteString(out)
					backtickCmdBuf.Reset()
				} else {
					currentArgBuf.WriteRune('$') // 如果不支持，则保留 $ ( )
					currentArgBuf.WriteRune('(')
					currentArgBuf.WriteString(backtickCmdBuf.String())
					currentArgBuf.WriteRune(')')
					backtickCmdBuf.Reset()
				}
				state = StateNormal
			} else {
				backtickCmdBuf.WriteRune(r)
			}
		}
	}

endLoop:
	// 循环结束后处理最后一个参数
	if currentArgBuf.Len() > 0 {
		args = append(args, currentArgBuf.String())
	}

	// 检查未闭合的引用或命令
	if state != StateNormal {
		return nil, fmt.Errorf("invalid command line string: unclosed %s", getStateName(state))
	}

	p.Position = pos
	return args, nil
}

// 辅助函数，用于获取状态的名称，方便错误信息
func getStateName(s argParseState) string {
	switch s {
	case StateNormal:
		return "normal"
	case StateEscaped:
		return "escape sequence"
	case StateSingleQuote:
		return "single quote"
	case StateDoubleQuote:
		return "double quote"
	case StateBacktick:
		return "backtick"
	case StateDollarCommand:
		return "dollar command"
	default:
		return "unknown state"
	}
}

func Parse(line string) ([]string, error) {
	return NewParser().Parse(line)
}
