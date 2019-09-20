package authz

import data.policies

default allow = false

allow {
    policy := policies[input.user]
    policy.effect == "allow"
}
