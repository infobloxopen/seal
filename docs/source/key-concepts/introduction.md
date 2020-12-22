# Key Concepts

SEAL has a couple of key concepts that are important to grasp.

* *policy* is a set of rules that specifies authorization decisions for resources.

* *rule* defines one singular action to apply to one specific operation depending on conditions.

* *subject* is an authenticated principal. This is an entity that has
  passed the authN step in the request process.

* *permission* is low level operation that is trying to be performed. In
  SEAL, we group permissions into verbs (aka roles). Permissions are what
  applications check in order to

* *verb* is an operation that a subject is trying to perform. Verbs can also
  be thought of as roles with the caveat that they should be evocative
  so that they are easy to read in a policy rule.

* *action* is a consequence of a policy rule decision. The default action,
  for example, could be deny. In opa terms, this is a decision.

* *resource-type* is a resource that is being secured. In many cases these are domain objects
  in the system under use. Sometimes, they are synthetic (made up) types that aren't
  stored anywhere but created in order to model authorization requests.

* *resource-family* is a group of resources that can be referenced together in policy rules.
  This allows policy rules to be more succinct if the overall policy needs to have access
  to several types.


From these basic concepts, SEAL allows users to create *action* rules that
describe an authorization policy. For example:

```bash
allow subject group foo to manage products.*;
```

In the above rule, subjects who are in the foo group can manage any types that
are in the `products` resource family. The verbs referenced in action
rules can also be defined. SEAL ships with some predefined verbs
and permissions to get you started.

# Subjects

SEAL is used to authorize someone against some resources. In this context, someone is
called a subject. A subject can be a user or group. To reference a user in a policy rule
you can use the "user" keyword. Likewise, to reference a group in a policy rule you can
use the "group" keyword.

```bash
allow subject user someone@acme.com to manage products.inventory;
```

or

```bash
allow subject group finance to manage accounts.*;
```

# Permissions

Permissions are string that define a type of authorization or consent. Permissions
can be defined as valid for a type or type family. For many applications that follow
the familiar CRUD model, it may make sense to define a permission for every type
of operation: create, read, update & delete. In kubernetes, there is a standard set
of operations:

> API request verb - API verbs like `get`, `list`, `create`, `update`, `patch`, `watch`, `delete`, and `deletecollection` are used for resource requests.
[Kubernets Authorization Overview](https://kubernetes.io/docs/reference/access-authn-authz/authorization/#review-your-request-attributes)

# Verbs

Verbs are actions that a policy author can reference. These verbs are used to group
permissions together much in the same way a role is used in traditional RBAC. The difference
is that the permissions that are referenced can themselves only be valid for specific types.

```bash
openapi: "3.0.0"
components:
  schemas:
    # global mapping of verbs to permission that can be referenced by types
    verbs:
      type: object
      x-seal-type: verbs
      x-seal-verbs:
        inspect:   [ "list", "watch" ]
        read:      [ "get", "list", "watch" ]
        use:       [ "update", "get", "list", "watch" ]
        manage:    [ "create", "delete", "update", "get", "list", "watch" ]
    # example resource types that references global verbs and custom verb
    products.inventory:
      type: object
      x-seal-verbs:
        inspect:   [ "#/components/schemas/verbs/x-seal-verbs/inspect" ]
        read:      [ "#/components/schemas/verbs/x-seal-verbs/read" ]
        use:       [ "#/components/schemas/verbs/x-seal-verbs/use" ]
        manage:    [ "#/components/schemas/verbs/x-seal-verbs/manage" ]
        provision: [ "provision", "deprovision" ]  # non-global permissions
    ...
```

# Actions

Actions are the results of policy rule decisions. In SEAL, you can reference actions by associating them with a resource type. A very common set of actions is defined below.

```bash
openapi: "3.0.0"
components:
  schemas:
    products.inventory:
      type: object
      x-seal-actions:
      - allow
      - deny
      x-seal-default-action: deny
```

Sometimes it is useful to allow actions to have parameters. For example, you may want to log a special log message if a particular action is taken.

```yaml
 openapi: "3.0.0"
 components:
    schemas:
      allow:
        type: object
        properties:
          log:
            type: string
        x-seal-type: action
      products.inventory:
        type: object
        x-seal-actions:
        - allow
        - deny
        x-seal-default-action: deny
```

These versions of allow & deny would permit the following syntax:

```bash
allow (log="my special rule") subject user someone@acme.com to manage products.inventory;
```

