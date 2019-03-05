//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package aux

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"idcos.io/cloud-act2/define"

	yaml "gopkg.in/yaml.v2"
	"idcos.io/cloud-act2/service/common"
)

type DefaultOutput struct {
	results []common.HostResultCallback
	start   time.Time
	noColor bool
}

func getLocalTime(d time.Time) string {
	timeLocation, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return d.Format("15:04:05")
	} else {
		return d.In(timeLocation).Format("15:04:05")
	}
}

func NewDefaultOutput(results []common.HostResultCallback,
	start time.Time) *DefaultOutput {
	return &DefaultOutput{
		results: results,
		start:   start,
	}
}

func (d *DefaultOutput) String() string {
	buffer := bytes.NewBufferString("")
	for i, result := range d.results {
		fmt.Println()

		builder := &Builder{}

		duration := time.Since(d.start)

		currentTime := getLocalTime(d.start)
		builder = builder.Append(fmt.Sprintf("[%d]\t%s\t%s\t%.2f(s)\t\t", i+1, result.HostIP, currentTime, duration.Seconds()))
		status := fmt.Sprintf("[%s]\t", result.Status)

		if define.Success == result.Status {
			if d.noColor {
				builder.Append(status)
			} else {
				builder.Green(status)
			}

			// builder.Append(result.HostIP)
			fmt.Fprintln(buffer, builder.String())
			fmt.Fprintln(buffer, result.Stdout)
			continue
		} else {
			if d.noColor {
				builder.Append(status)
			} else {
				builder.Red(status)
			}
			fmt.Fprintln(buffer, builder.String())
		}

		if result.Message != "" {
			fmt.Fprintf(buffer, "%s\n\n", result.Message)
		} else {
			fmt.Fprintf(buffer, "%s\n\n", result.Stderr)
		}
	}
	return buffer.String()
}

type JSONOutput struct {
	results []common.HostResultCallback
	start   time.Time
	output  *Output
}

func NewJSONOutput(results []common.HostResultCallback, start time.Time, output *Output) *JSONOutput {
	return &JSONOutput{
		results: results,
		start:   start,
		output:  output,
	}
}

func (j *JSONOutput) String() string {
	byts, err := json.MarshalIndent(j.results, "", "\t")
	if err != nil {
		j.output.Printf("marshal to json error %s", err)
		return ""
	}

	return string(byts)
}

type YAMLOutput struct {
	results []common.HostResultCallback
	start   time.Time
	output  *Output
}

func NewYAMLOutput(results []common.HostResultCallback,
	start time.Time, output *Output) *YAMLOutput {
	return &YAMLOutput{
		results: results,
		start:   start,
		output:  output,
	}
}

func (y *YAMLOutput) String() string {
	byts, err := yaml.Marshal(y.results)
	if err != nil {
		y.output.Printf("marshal to yaml error %s", err)
		return ""
	}

	return string(byts)
}

type Output struct {
	writer  io.Writer
	verbose bool
}

func NewOutput(writer io.Writer, verbose bool) *Output {
	return &Output{
		writer:  writer,
		verbose: verbose,
	}
}

func (o *Output) Printf(format string, args ...interface{}) {
	fmt.Fprintf(o.writer, format, args...)
}

func (o *Output) Verbose(format string, args ...interface{}) {
	if o.verbose {
		fmt.Fprintf(o.writer, format, args...)
	}
}
