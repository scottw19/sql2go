package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"strings"
)

func main() {
	var (
		inputPath   = flag.String("i", "", "Input filepath")
		outfilePath = flag.String("o", "", "Output filepath")
		includeJson = flag.Bool("j", false, "Include json struct tags")
	)
	flag.Parse()

	if *inputPath == "" {
		log.Fatal("No input filepath provided. Use -i.")
	}
	if strings.Index(*inputPath, ".sql") < 0 {
		log.Fatal("Input file must be an .sql file.")
	}

	sql2go(*inputPath, *outfilePath, *includeJson)
}

// meta struct
type strukt struct {
	name             string
	fields           []field
	longestFieldName int // used for putting spaces after field names
	longestTyp       int // used for putting spaces after field types for json tags
}
type field struct {
	name string
	typ  string
	jtag string
}

func sql2go(path, outfilePath string, includeJson bool) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var builder strings.Builder
	var line string
	var s *strukt
	for scanner.Scan() {
		line = scanner.Text()
		// lines containing CREATE TABLE are the start of a new struct
		if strings.Contains(line, "CREATE TABLE") {
			s = &strukt{}
			// CREATE TABLE public.access_tokens (
			tname := line[20 : len(line)-2]
			sname := snakeToCamelCase(tname)
			// Remove pluralization for table names
			// TODO: Handle tables that are singular but which end in "s" somehow?
			// e.g. ncis
			if sname[len(sname)-1] == 115 {
				sname = sname[0 : len(sname)-1]
			}
			s.name = sname
			continue
		}
		if s != nil && line == ");" {
			writeStrukt(&builder, *s)
			s = nil
			continue
		}
		if s != nil {
			line = strings.TrimSpace(line)
			lineParts := strings.Split(line, " ")
			if len(lineParts) < 2 {
				log.Fatal("wtf is this line", line)
			}
			colName := lineParts[0]

			// ignore lines which begin with CONSTRAINT
			// e.g. CONSTRAINT priority_positive CHECK ((priority >= 1))
			if colName == "CONSTRAINT" {
				continue
			}

			// new field
			f := field{
				name: snakeToCamelCase(colName),
			}
			if s.longestFieldName < len(f.name) {
				s.longestFieldName = len(f.name)
			}

			colType := lineParts[1]

			// remove trailing commas on field types if present
			if colType[len(colType)-1] == ',' {
				colType = colType[0 : len(colType)-1]
			}

			switch colType {
			case "bigint":
				f.typ = "int64"
			case "integer":
				f.typ = "int"
			case "smallint":
				f.typ = "int16"
			case "text[]":
				f.typ = "[]string"
			case "jsonb":
				f.typ = "json.RawMessage"
			case "boolean":
				f.typ = "bool"
			case "bytea":
				f.typ = "[]byte"

				// various types just dumping to string for now
			case "text", "date", "uuid", "interval", "\"char\"", "inet":
				f.typ = "string"
			default:
				if strings.Index(colType, "public.") == 0 {
					// any self-made types can just be dumped to string for now
					f.typ = "string"
					break
				}
				if strings.Index(colType, "numeric") == 0 {
					f.typ = "float32"
					break
				}
				if strings.Index(colType, "timestamp") == 0 {
					// dump timestamps to string for now
					f.typ = "string"
					break
				}
				if strings.Index(colType, "character") == 0 {
					f.typ = "string"
					break
				}
				log.Fatal("unhandled colType:", colType)
			}
			if includeJson {
				if s.longestTyp < len(f.typ) {
					s.longestTyp = len(f.typ)
				}
				f.jtag = colName
			}
			s.fields = append(s.fields, f)
		}
	}

	if outfilePath != "" {
		println("Writing to", outfilePath)
		err = os.WriteFile(outfilePath, []byte(builder.String()), 0660)
		if err != nil {
			log.Panicln("Unable to write outfile")
		}
	} else {
		println(builder.String())
	}
}

// snakeToCamelCase takes a snake_case string and
// converts it to CamelCase
func snakeToCamelCase(tname string) string {
	retval := []rune{}
	cap := true
	for _, b := range tname {
		if cap {
			// subtract 32 to capitalize
			b -= 32
			cap = false
		}
		if b == 95 {
			cap = true
			continue
		}
		retval = append(retval, b)
	}
	return string(retval)
}

func writeStrukt(b *strings.Builder, s strukt) error {
	// declaration
	b.WriteString("type " + s.name + " struct {\n")

	// fields
	for _, f := range s.fields {
		b.WriteString("  " + f.name)
		for i := len(f.name); i < s.longestFieldName; i++ {
			b.WriteByte(' ')
		}
		b.WriteString(" " + f.typ)
		if f.jtag != "" {
			for i := len(f.typ); i < s.longestTyp; i++ {
				b.WriteByte(' ')
			}
			b.WriteString(" `json:\"" + f.jtag + "\"`")
		}
		b.WriteByte('\n')
	}

	// end
	b.WriteString("}\n\n")
	return nil
}
