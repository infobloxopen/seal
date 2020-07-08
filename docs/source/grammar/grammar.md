

# Action Statements
Action statements governs the control of resources and follows this overall basic syntax:
```
<action>
    [with <action-modifier>]
    [to <action-object>]
[subject <subject-type> <subject> who]
<verb> <resource>
[where <condition>+]
```

The following is an example of the simplest action statement that allows everyone to resolve DNS requests:
```
allow resolve dns.request
```


## Action Phrase
The action phrase starts an action statement and is composed of an action, followed by an optional modifier and an optional object:
```
<action>
    [with <action-modifier>]
    [to <action-object>]
```

## Action
An action is the first word in the policy statement and specifies the action to be taken where a set of conditions are met.  Examples of actions:
```
allow | deny | redirect | drop
```

## Action Modifier
An optional action modifier may be specified on the action that begins with *with*.  Examples of action modifiers:
```
with log
```

## Action Object
When an action is taken, an optional action object can be specified in the action phrase that begins with *to*. Examples of action objects:
```
to 0.0.0.0/0
to university.edu
to data.good_domains
to feed.edu_domains
```

## Subject Phrase
Subject phrases are composed of the keyword *subject* followed by the subject type, the subject, and the keyword *to*. A subject phrase is optional in a policy statement.  The implicit subject is `subject group * to` and denotes everyone.  The syntax of subject phrases:
```
[subject <subject-type> <subject> to]
```

## Subject Type
The subject type defines the type of the subject attempting to take an action.  Examples of subject types include:
```
user | group
```

## Subject
The subject describes the person or group attempting to take an action. 

Examples of subject phrases include:
```
subject group students to
subject user robert@acme.com to
subject group fourth-graders to
subject group * to
```

## Verb
The verb describes the action that the subject is attempting to use.  

## Resource
The resource describes the resource that the subject is attempting to use.

## Where Phrase
A where phrase describes one or more conditions to be satisfied in the policy statement. 



# Examples
## 1. simple examples
```
allow subject group sysadmins who manage hosts.*
allow subject group dnsadmins who operate dns.*
allow subject group secadmins who operate firewalls.*
allow inspect dns.*
```

## 2. complex examples
```
# ==== dns section
[dns]
redirect to university.edu subject group students who resolve dns.request where req.domain in gambling_domains
redirect subject group students who resolve dns.request where dst.domain in redirect_domains

log when group students resolve dns.request where req.domain in edu_domains

allow subject group students who resolve dns.request where req.domain in edu_domains
allow subject user robert@acme.com who resolve dns.request where req.dest in good_sites  # customer defines good_sites

drop subject user pete.rose@chicago.il.us who resolve dns.request where req.domain matches *.vegas.nevada.us 
drop subject group students who resolve dns.request where dst.domain in bad_domains
 
# ==== firewall section
[firewalls]
allow subject group * who pass_thru firewall.endpoint where src.address == 10.1.2.4 and dst.address == 124.32.11.13
 
allow subject group * who pass_thru firewall.endpoint where src.address in 10.1/16
drop pass_thru firewall.endpoint where src.address == 10.1.1.123
 
allow subject group * who pass_thru firewall.endpoint where src.address in 10/8
drop pass_thru firewall.endpoint where src.address in 10.2/16
 
drop with log pass_thru firewall.endpoint where req.domain in hacker_domains
 
# deny fourth graders to connect to gambling sites
drop subject group fourth_graders who pass_thru firewall.endpoint where protocol == * and req.domain in gambling_sites
 
# allow ip to connect anywhere from a specific address on tcp
allow to 0.0.0.0/0 subject group * who pass_thru firewall.endpoint where src.address = 192.168.22.11:* and protocol == tcp
 
# allow anyone to connect to good_sites for any protocol
allow to 0.0.0.0/0 subject group * who pass_thru firewall.endpoint where dst.address in good_sites and protocol == *
 
# allow students to connect to customized good_domains
allow to good-domains subject group students who pass_thru firewall.endpoint where protocol == *
 
# ==== host section
[host]
notify to administrators subject host * who use memory where memory.used_percent > 80
notify to administrators subject user * who becomes user.root
```

# EBNF (Extended Backus Naur Form) Grammar
A policy statement consists of the following in EBNF grammar:
```
<policy-statement>   ::= <action> [to <action-object>] [<subject-phrase>] <verb> <resource> [<where-phrase>]
 
; ==== action-phrase dependencies section
<action>             ::= <action-char> <action>
<action-char>        ::= <letter> | <digit> | "_" | "-" 
 
<action-object>      ::= <action-object-char> <action-object>
<action-object-char> ::= <letter> | <digit> | "_" | "-" | "." | "/"
 
; ==== subject-phrase dependencies section
<subject-phrase>     ::= subject <subject-type> <subject> who
<subject>            ::= <subject-char> <subject>
<subject-char>       ::= <letter> | <digit> | "_" | "-" | "."
 
; ==== verb and resource dependencies section
<verb>               ::= <verb-char> <verb>
<verb-char>          ::= <letter> | <digit> | "_" | "-"
 
<resource>           ::= <resource-char> <resource>
<resource-char>      ::= <letter> | <digit> | "_" | "-" | "."
 
; ==== where-phrase dependencies section
<where-phrase>                ::= where <condition>+
<condition>                   ::= <conditional-or-expression>
```
