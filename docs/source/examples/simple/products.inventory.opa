
package products.inventory.seal
allow = true {
    `operators` in input.subject.groups
    input.verb == `use`
    re_match(`products.inventory`, input.type)
} where {
    ctx.id = "bar" and ctx.name = "foo"
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
