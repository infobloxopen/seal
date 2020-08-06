

# Action Statements
Action statements governs the control of resources and follows this overall basic syntax:
```
<action> [<action-object>+]
[subject <subject-type> <subject>]
to <verb> <resource>
[where <condition>+]
;
```

The following is an example of the simplest action statement that allows everyone to resolve DNS requests:
```
allow to resolve dns.request;
```


## Action Clause
The action clause starts an action statement and is composed of an action, followed by optional objects:
```
<action> [<action-object>+]
```

## Action
An action is the first word in the policy statement and specifies the action to be taken where a set of conditions are met.  Examples of actions:
```
allow | deny | redirect | redirect_to | drop
```

## Action Object
When an action is taken, optional action objects can be specified in the action clause.  Examples of action objects:
```
0.0.0.0/0
university.edu
data.good_domains
feed.edu_domains
```

## Subject Clause
Subject clauses are composed of the keyword *subject* followed by the subject type, the subject, and the keyword *to*. A subject clause is optional in a policy statement.  The implicit subject is `subject group * to` and denotes everyone.  The syntax of subject clauses:
```
[subject <subject-type> <subject>]
```

## Subject Type
The subject type defines the type of the subject attempting to take an action.  Examples of subject types include:
```
user | group
```

## Subject
The subject describes the person or group attempting to take an action. 

Examples of subject clauses include:
```
subject user robert@acme.com
subject group students
subject group fourth-graders
subject group *
```

## Verb
The verb describes the action that the subject is attempting to use.  

## Resource
The resource describes the resource that the subject is attempting to use.

## Where Clause
A where clause describes one or more conditions to be satisfied in the policy statement. 



# Examples
## 1. simple examples
```
allow subject group sysadmins to manage hosts.*;
allow subject group dnsadmins to operate dns.*;
allow subject group secadmins to operate firewalls.*;
allow to inspect dns.*;
```


## 2. complex examples
```
# ==== dns section
[dns]
redirect_to university.edu subject group students to resolve dns.request where req.domain in gambling_domains;
redirect subject group students to resolve dns.request where dst.domain in redirect_domains;

log when group students to resolve dns.request where req.domain in edu_domains;

allow subject group students to resolve dns.request where req.domain in edu_domains;
allow subject user robert@acme.com to resolve dns.request where req.dest in good_sites  # customer defines good_sites;

drop subject user pete.rose@chicago.il.us to resolve dns.request where req.domain matches *.vegas.nevada.us;
drop subject group students to resolve dns.request where dst.domain in bad_domains;
 
# ==== firewall section
[firewalls]
allow subject group * to pass_thru firewall.endpoint where src.address == 10.1.2.4 and dst.address == 124.32.11.13;
 
allow subject group * to pass_thru firewall.endpoint where src.address in 10.1/16;
drop pass_thru firewall.endpoint where src.address == 10.1.1.123;
 
allow subject group * to pass_thru firewall.endpoint where src.address in 10/8;
drop to pass_thru firewall.endpoint where src.address in 10.2/16;
 
drop to pass_thru firewall.endpoint where req.domain in hacker_domains;
 
# deny fourth graders to connect to gambling sites
drop subject group fourth_graders to pass_thru firewall.endpoint where protocol == * and req.domain in gambling_sites;
 
# allow ip to connect anywhere from a specific address on tcp
allow_to 0.0.0.0/0 subject group * to pass_thru firewall.endpoint where src.address = 192.168.22.11:* and protocol == tcp;
 
# allow anyone to connect to good_sites for any protocol
allow_to 0.0.0.0/0 subject group * to pass_thru firewall.endpoint where dst.address in good_sites and protocol == *;
 
# allow students to connect to customized good_domains
allow_to good-domains subject group students to pass_thru firewall.endpoint where protocol == *;
 
# ==== host section
[host]
notify_to administrators subject host * to use memory where memory.used_percent > 80;
notify_to administrators subject user * to becomes user.root;
```

# Context Stanzas

Writing action statements can become repetitive. In some cases you may want to repeat the same *subject* or *where*
specifications repeatedly. To simplify writing groups of action statements with similar criteria
you can use a context stanza.

```
context {
    [ [<subject-clause>] [<where-clause>] ];
    ...
} [to [<verb>]] [<resource>] {
    [<action-statement>];
    ...
};
```

The first block after the context keyword allows for subject and where clauses to be specified. These clauses are
command separated. A semicolon terminates all the clauses that are logically ORed. You can repeat this pattern
as many times as you like. After this block comes an optional VERB and TYPE specification. This specification
will implicitly be applied to ever action statement in the following block. Next comes an action block. Here
you can add action statements.

The overall behavior is that for each statement in the action block the optional verb and type are applied. Then,
the devived action statement is repeated for every statement in the context block. Lets look at an example.

```
context {
    subject group engineering where req.tags["dept"] == "engineering";
    subject group everyone where req.scope == "public";
} to manage {
    allow products-family;
    allow inventory-family;
};
```

The net effect is that the backend compilers will interpret this block as follows.
```
allow subject group engineering to manage products-family where req.tags["dept"] == "engineering";
allow subject group all to manage products-family where req.scope == "public";

allow subject group engineering manage inventory-family where req.tags["dept"] == "engineering";
allow subject group all to manage inventory-family where req.scope == "public";
```

Context blocks can also be nested.
```
context {
    where req.tenant == "acme.com";
} {
    context {
        subject group engineering where req.tags["dept"] == "engineering";
        subject group everyone where req.scope == "public";
    } to manage {
        allow products-family;
        allow inventory-family;
    }
}
```

# EBNF (Extended Backus Naur Form) Grammar
A policy statement consists of the following in EBNF grammar:
```
<policy-statement>   ::= <action-statement> | <context-statement>
```

An action policy statement consists of the following in EBNF grammar:
```
<action-statement>   ::= <action> [<action-object> ...] [<subject-clause>] to <verb> <resource> [<where-clause>];
 
; ==== action-clause dependencies section
<action>             ::= <action-char> <action>
<action-char>        ::= <letter> | <digit> | "_" | "-" 
 
<action-object>      ::= <action-object-char> <action-object>
<action-object-char> ::= <letter> | <digit> | "_" | "-" | "." | "/"
 
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

A context policy statement consists of the following in EBNF grammar:
```
<context-statement>  ::= context "{" [<context-principals>] "}" [to [<verb>]] [<resource>] "{" [<action-statements>] "}";

<context-principals> ::= <context-principal>+
<context-principal>  ::= [<subject-clause>] [<where-clause>];

<action-statements>  ::= <action-statement>+
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

<not-expression>              ::= <not-expression> | not <primary>

<primary>                     ::= <literal> | ( <expression> ) | <field-access>

<field-access>                ::= <identifier> . <identifier>

<identifier>                  ::= <identifier-char> <identifier>
<identifier-char>             ::= <letter> | <digit> | "_"

; ==== common dependencies section
<literal>                     ::= <integer> | '"' <quoted> '"'

<integer>                     ::= <integer> | <digit> <integer>

<quoted>                      ::= <quoted> | <quoted-char> <quoted>
<quoted-char>                 ::= <letter> | <digit> | "_" | "-" | "." | "@" | "/" | "+" | "*"

<letter>                      ::= "A" | "B" | "C" | "D" | "E" | "F" | "G" | "H" | "I" | "J" | "K" | "L" | "M" | "N" | "O" | "P" | "Q" | "R" | "S" | "T" | "U" | "V" | "W" | "X" | "Y" | "Z" | "a" | "b" | "c" | "d" | "e" | "f" | "g" | "h" | "i" | "j" | "k" | "l" | "m" | "n" | "o" | "p" | "q" | "r" | "s" | "t" | "u" | "v" | "w" | "x" | "y" | "z"

<digit>                       ::= "0" | "1" | "2" | "3" | "4" | "5" | "6" | "7" | "8" | "9"

```
