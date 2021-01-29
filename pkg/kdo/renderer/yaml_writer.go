package renderer

import "io"

// YamlWriter -
type YamlWriter struct {
	state  int
	Writer io.Writer
}

const skipWhitespace = 0
const separator = 1
const endSeperator = 2
const normal = 3

func (w *YamlWriter) Write(data []byte) (int, error) {
	if w.state != normal {
		for i, b := range data {
			switch w.state {
			case skipWhitespace:
				switch b {
				case ' ':
				case '\n':
				case '-':
					w.state = separator
				default:
					w.state = normal
				}
			case separator:
				switch b {
				case '-':
				case '\n':
					w.state = endSeperator
				default:
					w.state = normal
				}
			case endSeperator:
				w.state = normal
			}
			if w.state == normal {
				w.Writer.Write([]byte("---\n"))
				return w.Writer.Write(data[i:])
			}
		}
		return 0, nil
	}
	return w.Writer.Write(data)
}
