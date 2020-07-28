
package products.inventory.seal
allow = true {
    input.subject.user == `somebody`
    input.verb == `manage`
    re_match(`products.inventory`, input.type)
}
allow = true {
    `everyone` in input.subject.groups
    input.verb == `manage`
    re_match(`products.inventory`, input.type)
}
