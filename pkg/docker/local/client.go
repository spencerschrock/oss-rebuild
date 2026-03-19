// Copyright 2025 Google LLC
// SPDX-License-Identifier: Apache-2.0

package local

import (
	"context"
	"io"
)

// Client interface abstracts Docker commands.
type Client interface {
	Build(ctx context.Context, tag string, dockerfile io.Reader, output io.Writer) error
	Run(ctx context.Context, image string, output io.Writer, args []string) error
	Save(ctx context.Context, image string, outputPath string) error
	Rmi(ctx context.Context, image string) error
}

type realClient struct {
	cmdExecutor CommandExecutor
	dockerCmd   string
}

// NewRealClient creates a default Docker Client implementation.
func NewRealClient(cmdExecutor CommandExecutor, dockerCmd string) Client {
	return &realClient{
		cmdExecutor: cmdExecutor,
		dockerCmd:   dockerCmd,
	}
}

func (c *realClient) Build(ctx context.Context, tag string, dockerfile io.Reader, output io.Writer) error {
	buildArgs := []string{"buildx", "build", "-t", tag, "-"}
	return c.cmdExecutor.Execute(ctx, CommandOptions{
		Input:  dockerfile,
		Output: output,
	}, c.dockerCmd, buildArgs...)
}

func (c *realClient) Run(ctx context.Context, image string, output io.Writer, args []string) error {
	runArgs := append([]string{"run"}, args...)
	if image != "" {
		runArgs = append(runArgs, image)
	}
	return c.cmdExecutor.Execute(ctx, CommandOptions{
		Output: output,
	}, c.dockerCmd, runArgs...)
}

func (c *realClient) Save(ctx context.Context, image string, outputPath string) error {
	return c.cmdExecutor.Execute(ctx, CommandOptions{}, c.dockerCmd, "save", "-o", outputPath, image)
}

func (c *realClient) Rmi(ctx context.Context, image string) error {
	return c.cmdExecutor.Execute(ctx, CommandOptions{}, c.dockerCmd, "rmi", image)
}
