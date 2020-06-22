/*
Copyright (c) 2018 VMware, Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tags

import (
	"context"
	"flag"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/vmware/govmomi/govc/cli"
	"github.com/vmware/govmomi/govc/flags"
	"github.com/vmware/govmomi/vapi/tags"
)

type ls struct {
	*flags.ClientFlag
	*flags.OutputFlag
	c string
}

func init() {
	cli.Register("tags.ls", &ls{})
}

func (cmd *ls) Register(ctx context.Context, f *flag.FlagSet) {
	cmd.ClientFlag, ctx = flags.NewClientFlag(ctx)
	cmd.OutputFlag, ctx = flags.NewOutputFlag(ctx)
	cmd.ClientFlag.Register(ctx, f)
	cmd.OutputFlag.Register(ctx, f)
	f.StringVar(&cmd.c, "c", "", "Category name")
}

func (cmd *ls) Process(ctx context.Context) error {
	if err := cmd.ClientFlag.Process(ctx); err != nil {
		return err
	}
	return nil
}

func (cmd *ls) Description() string {
	return `List tags.

Examples:
  govc tags.ls
  govc tags.ls -c k8s-zone
  govc tags.ls -json | jq .
  govc tags.ls -c k8s-region -json | jq .`
}

type lsResult []tags.Tag

func (r lsResult) Write(w io.Writer) error {
	tw := tabwriter.NewWriter(w, 2, 0, 2, ' ', 0)

	for i := range r {
		fmt.Fprintf(tw, "%s\t%s\n", r[i].Name, r[i].CategoryID)
	}

	return tw.Flush()
}

func (cmd *ls) Run(ctx context.Context, f *flag.FlagSet) error {
	c, err := cmd.RestClient()
	if err != nil {
		return err
	}

	m := tags.NewManager(c)
	var res lsResult

	if cmd.c == "" {
		res, err = m.GetTags(ctx)
	} else {
		res, err = m.GetTagsForCategory(ctx, cmd.c)
	}

	if err != nil {
		return err
	}

	categories, err := m.GetCategories(ctx)
	if err != nil {
		return err
	}
	cats := make(map[string]tags.Category)
	for _, category := range categories {
		cats[category.ID] = category
	}

	for i, tag := range res {
		res[i].CategoryID = cats[tag.CategoryID].Name
	}

	return cmd.WriteResult(res)
}
