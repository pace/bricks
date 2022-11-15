// Copyright © 2019 by PACE Telematics GmbH. All rights reserved.
// Created at 2019/03/29 by Florian Hübsch

package oauth2

import "testing"

func TestIsIncludedIn(t *testing.T) {
	requiredScope := Scope("a:read b:write:x c f*")

	tcs := []struct {
		t  Scope
		ex bool
	}{
		{Scope("a:read"), false},
		{Scope("b:write:x"), false},
		{Scope("c"), false},
		{Scope("a"), false},
		{Scope("b:write"), false},
		{Scope("c a:read"), false},
		{Scope("b:write:x c a:read f*"), true},
		{Scope("a:read b:write:x c f* d"), true},
		{Scope("foo a:read c f* bar b:write:x"), true},
		{Scope("a:read c fg bar b:write:x"), false},
	}

	for _, tc := range tcs {
		got := requiredScope.IsIncludedIn(tc.t)
		if got != tc.ex {
			t.Errorf("Expected %q to be included in %q", requiredScope, tc.t)
		}
	}
}

func TestScope_Add(t *testing.T) {
	type args struct {
		scope string
	}
	tests := []struct {
		name string
		s    Scope
		args args
		want Scope
	}{
		{"add one", Scope("a:read"), args{"b:write"}, Scope("a:read b:write")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.Add(tt.args.scope); got != tt.want {
				t.Errorf("Add() = %v, want %v", got, tt.want)
			}
		})
	}
}
