// Copyright 2020 the Pinniped contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1alpha1 "go.pinniped.dev/generated/1.18/apis/login/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeTokenCredentialRequests implements TokenCredentialRequestInterface
type FakeTokenCredentialRequests struct {
	Fake *FakeLoginV1alpha1
	ns   string
}

var tokencredentialrequestsResource = schema.GroupVersionResource{Group: "login.pinniped.dev", Version: "v1alpha1", Resource: "tokencredentialrequests"}

var tokencredentialrequestsKind = schema.GroupVersionKind{Group: "login.pinniped.dev", Version: "v1alpha1", Kind: "TokenCredentialRequest"}

// Get takes name of the tokenCredentialRequest, and returns the corresponding tokenCredentialRequest object, and an error if there is any.
func (c *FakeTokenCredentialRequests) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.TokenCredentialRequest, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(tokencredentialrequestsResource, c.ns, name), &v1alpha1.TokenCredentialRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.TokenCredentialRequest), err
}

// List takes label and field selectors, and returns the list of TokenCredentialRequests that match those selectors.
func (c *FakeTokenCredentialRequests) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.TokenCredentialRequestList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(tokencredentialrequestsResource, tokencredentialrequestsKind, c.ns, opts), &v1alpha1.TokenCredentialRequestList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.TokenCredentialRequestList{ListMeta: obj.(*v1alpha1.TokenCredentialRequestList).ListMeta}
	for _, item := range obj.(*v1alpha1.TokenCredentialRequestList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested tokenCredentialRequests.
func (c *FakeTokenCredentialRequests) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(tokencredentialrequestsResource, c.ns, opts))

}

// Create takes the representation of a tokenCredentialRequest and creates it.  Returns the server's representation of the tokenCredentialRequest, and an error, if there is any.
func (c *FakeTokenCredentialRequests) Create(ctx context.Context, tokenCredentialRequest *v1alpha1.TokenCredentialRequest, opts v1.CreateOptions) (result *v1alpha1.TokenCredentialRequest, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(tokencredentialrequestsResource, c.ns, tokenCredentialRequest), &v1alpha1.TokenCredentialRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.TokenCredentialRequest), err
}

// Update takes the representation of a tokenCredentialRequest and updates it. Returns the server's representation of the tokenCredentialRequest, and an error, if there is any.
func (c *FakeTokenCredentialRequests) Update(ctx context.Context, tokenCredentialRequest *v1alpha1.TokenCredentialRequest, opts v1.UpdateOptions) (result *v1alpha1.TokenCredentialRequest, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(tokencredentialrequestsResource, c.ns, tokenCredentialRequest), &v1alpha1.TokenCredentialRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.TokenCredentialRequest), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeTokenCredentialRequests) UpdateStatus(ctx context.Context, tokenCredentialRequest *v1alpha1.TokenCredentialRequest, opts v1.UpdateOptions) (*v1alpha1.TokenCredentialRequest, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(tokencredentialrequestsResource, "status", c.ns, tokenCredentialRequest), &v1alpha1.TokenCredentialRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.TokenCredentialRequest), err
}

// Delete takes name of the tokenCredentialRequest and deletes it. Returns an error if one occurs.
func (c *FakeTokenCredentialRequests) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(tokencredentialrequestsResource, c.ns, name), &v1alpha1.TokenCredentialRequest{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeTokenCredentialRequests) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(tokencredentialrequestsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.TokenCredentialRequestList{})
	return err
}

// Patch applies the patch and returns the patched tokenCredentialRequest.
func (c *FakeTokenCredentialRequests) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.TokenCredentialRequest, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(tokencredentialrequestsResource, c.ns, name, pt, data, subresources...), &v1alpha1.TokenCredentialRequest{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.TokenCredentialRequest), err
}
