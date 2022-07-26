

# Action Rules
An action rule defines one singular action to apply to one specific operation depending on conditions.  Each rule follows this overall basic syntax:
```
<action> [( <action-property>+ )]
[subject <subject-type> <subject>]
to <verb> <resource>
[where <condition>+]
;
```

The following is an example of the simplest action rule that allows everyone to view all products:
```
allow to inspect products.inventory;
```

## Action Clause
The action clause starts an action rule and is composed of an action, followed by optional properties:
```
<action> [( <action-property>+ )]
```

## Action
An action is the first word in the action rule and specifies the action to be taken where a set of conditions are met.  Examples of actions:
```
allow | deny | redirect | drop
```

## Action Property
When an action is taken, optional action properties can be specified in the action clause.  Examples of action properties:
```
to="911"
to=$list["name=customer_support"]
```

## Subject Clause
Subject clauses are composed of the keyword *subject* followed by the subject type, the subject. A subject clause is optional in a policy rule.  The implicit subject denotes everyone.  The syntax of subject clauses:
```
[subject <subject-type> <subject>]
```

## Subject Type
The subject type defines the type of the subject attempting to take an action.  Examples of subject types include:
```
user | group
```

## Subject
The subject is an entity (person or application or system or service) that can make a request for an action or operation on a resource.

Examples of subject clauses include:
```
subject user robert@acme.com
subject group students
subject group fourth_graders
```

## Verb
The verb describes the action that the subject is attempting to use and is preceded by the keyword *to*.

## Resource
The resource describes the resource that the subject is attempting to use.

## Where Clause
A where clause describes one or more conditions to be satisfied in the policy rule.



# Examples
## 1. simple examples
```
allow to inspect products.inventory;
allow subject group operators to use products.inventory;
allow subject group admins to manage products.inventory;
allow subject user cto@acme.com to manage products.inventory;

allow subject group hr to manage company.personnel;
allow subject group finance to manage accounts.*;
```


## 2. complex examples
```
# ==== products section
[products]
allow subject group patissiers to manage products.inventory where ctx.tag["department"] == "bakery" and ctx.name == "cheesecake";

deny subject user pete.rose@chicago.il.us to buy products.inventory where ctx.genre == "gambling";

deny (log="true") subject group minors to buy products.inventory where ctx.sku in $threat.feed["over_21_skus"];

# ==== company section
[company]
redirect (to=$list["name=customer_support"], log="true") to seek company.help;
```

# Context Stanzas

Writing action rules can become repetitive. In some cases you may want to repeat the same *subject* or *where*
specifications repeatedly. To simplify writing groups of action rules with similar criteria
you can use a context stanza.

```
context {
    [ [<subject-clause>] [<where-clause>] ];
    ...
} [to [<verb>]] [<resource>] {
    [<action-rule>];
    ...
};
```

The first block after the context keyword allows for subject and where clauses to be specified. These clauses are
command separated. A semicolon terminates all the clauses that are logically ORed. You can repeat this pattern
as many times as you like. After this block comes an optional VERB and TYPE specification. This specification
will implicitly be applied to every action rule in the following block. Next comes an action block. Here
you can add action rules.

The overall behavior is that for each rule in the action block the optional verb and type are applied. Then,
the derived action rule is repeated for every rule in the context block. Lets look at an example.

```
context {
    subject group engineering where ctx.tags["dept"] == "engineering";
    subject group everyone where ctx.scope == "public";
} to manage {
    allow products.*;
    allow inventory.*;
};
```

The net effect is that the backend compilers will interpret this block as follows.
```
allow subject group engineering to manage products.* where ctx.tags["dept"] == "engineering";
allow subject group all to manage products.* where ctx.scope == "public";

allow subject group engineering manage inventory.* where ctx.tags["dept"] == "engineering";
allow subject group all to manage inventory.* where ctx.scope == "public";
```

Context blocks can also be nested.
```
context {
    where ctx.tenant == "acme.com";
} {
    context {
        subject group engineering where ctx.tags["dept"] == "engineering";
        subject group everyone where ctx.scope == "public";
    } to manage {
        allow products.*;
        allow inventory.*;
    }
}
```

# EBNF (Extended Backus Naur Form) Grammar
A policy consists of the following in EBNF grammar:
```
<policy>             ::= <policy-rule>+
<policy-rule>        ::= <action-rule> | <context-stanza>
```

A policy action rule consists of the following in EBNF grammar:
```
<action-rule>        ::= <action> [<action-property> ...] [<subject-clause>] to <verb> <resource> [<where-clause>];

; ==== action-clause dependencies section
<action>             ::= <action-char> <action>
<action-char>        ::= <letter> | <digit> | "_" | "-"

<action-property>      ::= <action-property-char> <action-property>
<action-property-char> ::= <letter> | <digit> | "_" | "-" | "." | "/"

; ==== subject-clause dependencies section
<subject-clause>     ::= subject <subject-type> <subject>
<subject>            ::= <subject-char> <subject>
<subject-char>       ::= <letter> | <digit> | "_" | "-" | "."

; ==== verb and resource dependencies section
<verb>               ::= <verb-char> <verb>
<verb-char>          ::= <letter> | <digit> | "_" | "-"

<resource>           ::= <resource-char> <resource>
<resource-char>      ::= <letter> | <digit> | "_" | "-" | "."
```

A policy context stanza consists of the following in EBNF grammar:
```
<context-stanza>     ::= context "{" [<context-principals>] "}" [to [<verb>]] [<resource>] "{" [<action-rules>] "}";

<context-principals> ::= <context-principal>+
<context-principal>  ::= [<subject-clause>] [<where-clause>];

<action-rules>       ::= <action-rule>+
```

Where Clause consists of the following in EBNF grammar:
```
<where-clause>                ::= where <condition>+
<condition>                   ::= <conditional-or-expression>

<conditional-or-expression>   ::= <conditional-and-expression> | <conditional-or-expression> or <conditional-and-expression>

<conditional-and-expression>  ::= <equality-expression> | <conditional-and-expression> and <equality-expression>

<equality-expression>         ::= <relational-expression>
                                | <equality-expression> == <relational-expression>
                                | <equality-expression> != <relational-expression>

<relational-expression>       ::= <not-expression>
                                | <relational-expression> < <not-expression>
                                | <relational-expression> > <not-expression>
                                | <relational-expression> <= <not-expression>
                                | <relational-expression> >= <not-expression>
                                | <relational-expression> in <field-access>
                                | <relational-expression> in <array-literal>

<not-expression>              ::= <not-expression> | not <primary>

<primary>                     ::= <literal> | ( <expression> ) | <field-access>

<field-access>                ::= <identifier> . <identifier>

<identifier>                  ::= <identifier-char> <identifier>
<identifier-char>             ::= <letter> | <digit> | "_"

; ==== common dependencies section
<array-literal>               ::= "[" <array-items> "]"
<array-items>                 ::= <literal> | <array-items> "," <literal>

<literal>                     ::= <integer> | '"' <quoted> '"'

<integer>                     ::= <integer> | <digit> <integer>

<quoted>                      ::= <quoted> | <quoted-char> <quoted>
<quoted-char>                 ::= <letter> | <digit> | "_" | "-" | "." | "@" | "/" | "+" | "*"

<letter>                      ::= "A" | "B" | "C" | "D" | "E" | "F" | "G" | "H" | "I" | "J" | "K" | "L" | "M" | "N" | "O" | "P" | "Q" | "R" | "S" | "T" | "U" | "V" | "W" | "X" | "Y" | "Z" | "a" | "b" | "c" | "d" | "e" | "f" | "g" | "h" | "i" | "j" | "k" | "l" | "m" | "n" | "o" | "p" | "q" | "r" | "s" | "t" | "u" | "v" | "w" | "x" | "y" | "z"

<digit>                       ::= "0" | "1" | "2" | "3" | "4" | "5" | "6" | "7" | "8" | "9"

```
