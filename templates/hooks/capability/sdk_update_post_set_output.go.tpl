	if resp.Update.Errors != nil {
		respErrors := []error{}
		for _, respError := range resp.Update.Errors {
			msg := fmt.Sprintf("%s: %v", aws.ToString(respError.ErrorMessage), respError.ResourceIds)
			respErrors = append(respErrors, &smithy.GenericAPIError{
				Code: string(respError.ErrorCode),
				Message: msg,
			})
		}
		if len(respErrors) > 0 {
			return nil, fmt.Errorf("update failed with errors: %v", errors.Join(respErrors...))
		}
	}
