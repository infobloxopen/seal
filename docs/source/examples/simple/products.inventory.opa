
package products.inventory.seal
allow = true {
    `everyone` in input.subject.groups
    input.verb == `inspect`
    re_match(`products.inventory`, input.type)
}
allow = true {
    `operators` in input.subject.groups
    input.verb == `use`
    re_match(`products.inventory`, input.type)
}
allow = true {
    `admins` in input.subject.groups
    input.verb == `manage`
    re_match(`products.inventory`, input.type)
}
allow = true {
    input.subject.user == `cto@acme.com`
    input.verb == `manage`
    re_match(`products.inventory`, input.type)
}
