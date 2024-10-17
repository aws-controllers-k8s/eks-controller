
	if identifier.AdditionalKeys == nil {
		return ackerrors.MissingNameIdentifier
	}
	f0, f0ok := identifier.AdditionalKeys["identityProviderConfigName"]
	if f0ok {
		r.ko.Spec.OIDC = &svcapitypes.OIDCIdentityProviderConfigRequest{
			IdentityProviderConfigName: &f0,
		}
	} else {
		return ackerrors.MissingNameIdentifier
	}
