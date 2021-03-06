// Copyright © 2019 The Knative Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package apiserver

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"knative.dev/client/pkg/eventing/legacysources/v1alpha1"
	"knative.dev/client/pkg/kn/commands"
	"knative.dev/client/pkg/kn/commands/flags"
)

// NewAPIServerCreateCommand for creating source
func NewAPIServerCreateCommand(p *commands.KnParams) *cobra.Command {
	var updateFlags APIServerSourceUpdateFlags
	var sinkFlags flags.SinkFlags

	cmd := &cobra.Command{
		Use:   "create NAME --resource RESOURCE --service-account ACCOUNTNAME --sink SINK --mode MODE",
		Short: "Create an ApiServer source.",
		Example: `
  # Create an ApiServerSource 'k8sevents' which consumes Kubernetes events and sends message to service 'mysvc' as a cloudevent
  kn source apiserver create k8sevents --resource Event --service-account myaccountname --sink svc:mysvc`,

		RunE: func(cmd *cobra.Command, args []string) (err error) {
			if len(args) != 1 {
				return errors.New("requires the name of the source to create as single argument")
			}
			name := args[0]

			// get client
			apiSourceClient, err := newAPIServerSourceClient(p, cmd)
			if err != nil {
				return err
			}

			namespace := apiSourceClient.Namespace()

			dynamicClient, err := p.NewDynamicClient(namespace)
			if err != nil {
				return err
			}
			objectRef, err := sinkFlags.ResolveSink(dynamicClient, namespace)
			if err != nil {
				return fmt.Errorf(
					"cannot create ApiServerSource '%s' in namespace '%s' "+
						"because: %s", name, namespace, err)
			}

			// create
			resources, err := updateFlags.GetAPIServerResourceArray()
			if err != nil {
				return err
			}

			err = apiSourceClient.CreateAPIServerSource(
				v1alpha1.NewAPIServerSourceBuilder(name).
					ServiceAccount(updateFlags.ServiceAccountName).
					Mode(updateFlags.Mode).
					Resources(*resources).
					Sink(objectRef).
					Build())

			if err != nil {
				return fmt.Errorf(
					"cannot create ApiServerSource '%s' in namespace '%s' "+
						"because: %s", name, namespace, err)
			}

			if err == nil {
				fmt.Fprintf(cmd.OutOrStdout(), "ApiServer source '%s' created in namespace '%s'.\n", args[0], namespace)
			}

			return err
		},
	}
	commands.AddNamespaceFlags(cmd.Flags(), false)
	updateFlags.Add(cmd)
	sinkFlags.Add(cmd)
	cmd.MarkFlagRequired("resource")
	cmd.MarkFlagRequired("sink")
	return cmd
}
