package rest

type Hook func(trigger, resource string, user, r Resource) error

func nopHook(trigger, resource string, user, r Resource) error { return nil }
