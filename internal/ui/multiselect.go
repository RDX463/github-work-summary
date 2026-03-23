package ui

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

var ErrSelectionCancelled = errors.New("selection cancelled")

// SelectOption is one selectable entry in the checkbox list.
type SelectOption struct {
	Label string
	Value string
}

// MultiSelectCheckboxes runs a simple interactive checkbox selector in the terminal.
func MultiSelectCheckboxes(in io.Reader, out io.Writer, title string, options []SelectOption) ([]SelectOption, error) {
	if len(options) == 0 {
		return nil, fmt.Errorf("no options available")
	}

	reader := bufio.NewReader(in)
	selected := make(map[int]bool)

	for {
		renderOptions(out, title, options, selected)

		line, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil, ErrSelectionCancelled
			}
			return nil, fmt.Errorf("failed to read selection input: %w", err)
		}

		input := strings.TrimSpace(strings.ToLower(line))
		switch input {
		case "":
			continue
		case "d", "done":
			selectedOptions := collectSelected(options, selected)
			if len(selectedOptions) == 0 {
				fmt.Fprintln(out, "Select at least one repository before finishing.")
				continue
			}
			return selectedOptions, nil
		case "a", "all":
			for i := range options {
				selected[i] = true
			}
			continue
		case "n", "none":
			clear(selected)
			continue
		case "q", "quit", "cancel":
			return nil, ErrSelectionCancelled
		default:
			indices, err := parseToggleInput(input, len(options))
			if err != nil {
				fmt.Fprintf(out, "Invalid input: %v\n", err)
				continue
			}
			for _, idx := range indices {
				selected[idx] = !selected[idx]
			}
		}
	}
}

func renderOptions(out io.Writer, title string, options []SelectOption, selected map[int]bool) {
	fmt.Fprintln(out)
	fmt.Fprintln(out, title)
	fmt.Fprintln(out, "Toggle: 1 2 3 or 1,2,3 or ranges 2-5")
	fmt.Fprintln(out, "Commands: a=all, n=none, d=done, q=quit")
	fmt.Fprintln(out, strings.Repeat("-", 70))

	for i, option := range options {
		mark := " "
		if selected[i] {
			mark = "x"
		}
		fmt.Fprintf(out, "%3d. [%s] %s\n", i+1, mark, option.Label)
	}
	fmt.Fprint(out, "\nSelection > ")
}

func parseToggleInput(input string, max int) ([]int, error) {
	normalized := strings.ReplaceAll(input, ",", " ")
	fields := strings.Fields(normalized)
	if len(fields) == 0 {
		return nil, fmt.Errorf("empty selection")
	}

	indices := make(map[int]struct{})
	for _, field := range fields {
		if strings.Contains(field, "-") {
			parts := strings.SplitN(field, "-", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid range %q", field)
			}
			start, err := parseIndex(parts[0], max)
			if err != nil {
				return nil, err
			}
			end, err := parseIndex(parts[1], max)
			if err != nil {
				return nil, err
			}
			if start > end {
				start, end = end, start
			}
			for i := start; i <= end; i++ {
				indices[i] = struct{}{}
			}
			continue
		}

		idx, err := parseIndex(field, max)
		if err != nil {
			return nil, err
		}
		indices[idx] = struct{}{}
	}

	result := make([]int, 0, len(indices))
	for idx := range indices {
		result = append(result, idx)
	}
	sort.Ints(result)
	return result, nil
}

func parseIndex(raw string, max int) (int, error) {
	n, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0, fmt.Errorf("invalid number %q", raw)
	}
	if n < 1 || n > max {
		return 0, fmt.Errorf("number %d out of range (1-%d)", n, max)
	}
	return n - 1, nil
}

func collectSelected(options []SelectOption, selected map[int]bool) []SelectOption {
	result := make([]SelectOption, 0, len(selected))
	for i, option := range options {
		if selected[i] {
			result = append(result, option)
		}
	}
	return result
}
