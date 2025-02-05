	identityProviderConfigType := IdentityProviderConfigType
	input.IdentityProviderConfig = &svcsdktypes.IdentityProviderConfig{
		Name: r.ko.Spec.OIDC.IdentityProviderConfigName,
		Type: &identityProviderConfigType,
	}