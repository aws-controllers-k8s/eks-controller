
	if f0, f0ok := fields["identityProviderConfigName"]; f0ok {
		r.ko.Spec.OIDC = &svcapitypes.OIDCIdentityProviderConfigRequest{
			IdentityProviderConfigName: &f0,
		}
	} else {
		return ackerrors.MissingNameIdentifier
	}
