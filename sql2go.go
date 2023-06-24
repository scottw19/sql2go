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

func sql2go(path, outfilePath string, includeJson bool) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var builder strings.Builder
	var line string
	var inStruct bool
	for scanner.Scan() {
		line = scanner.Text()
		// lines containing CREATE TABLE are the start of a new struct
		if strings.Contains(line, "CREATE TABLE") {
			inStruct = true
			// CREATE TABLE public.access_tokens (
			tname := line[20 : len(line)-2]
			sname := snakeToCamelCase(tname)
			// Remove pluralization for table names
			// TODO: Handle tables that are singular but which end in "s" somehow?
			// e.g. ncis
			if sname[len(sname)-1] == 115 {
				sname = sname[0 : len(sname)-1]
			}
			builder.WriteString("type " + sname + " struct {\n")
			continue
		}
		if inStruct && line == ");" {
			builder.WriteByte('}')
			builder.WriteString("\n\n")
			inStruct = false
		}
		if inStruct {
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

			colType := lineParts[1]

			// remove trailing commas on field types if present
			if colType[len(colType)-1] == ',' {
				colType = colType[0 : len(colType)-1]
			}

			fName := snakeToCamelCase(colName)
			builder.WriteString("  " + fName + " ") // indent

			switch colType {
			case "bigint":
				builder.WriteString("int64")
			case "integer":
				builder.WriteString("int")
			case "smallint":
				builder.WriteString("int16")
			case "text[]":
				builder.WriteString("[]string")
			case "jsonb":
				builder.WriteString("json.RawMessage")
			case "boolean":
				builder.WriteString("bool")
			case "bytea":
				builder.WriteString("[]byte")

				// various types just dumping to string for now
			case "text", "date", "uuid", "interval", "\"char\"", "inet":
				builder.WriteString("string")
			default:
				if strings.Index(colType, "public.") == 0 {
					// any self-made types can just be dumped to string for now
					builder.WriteString("string")
					break
				}
				if strings.Index(colType, "numeric") == 0 {
					builder.WriteString("float32")
					break
				}
				if strings.Index(colType, "timestamp") == 0 {
					// dump timestamps to string for now
					builder.WriteString("string")
					break
				}
				if strings.Index(colType, "character") == 0 {
					builder.WriteString("string")
					break
				}
				log.Fatal("unhandled colType:", colType)
			}
			if includeJson {
				builder.WriteString(" `json:\"" + colName + "\"`")
			}
			builder.WriteByte('\n')
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
