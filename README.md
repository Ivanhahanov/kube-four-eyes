# Kube four eyes
> Kubernetes is a machine for machines

## Story
In a perfect world, nothing breaks. For security reasons, you have decided to avoid human intervention in kubernetes. Now, kubernetes can only be deployed via git. The Flux operator itself rolls back failed changes, and clusters run on Talos OS and are deployed using IaC. Everything is perfect, developers write their own code and testers work in an ethereal environment. The business is happy. 

But at one point something went wrong, somewhere they forgot to indent or didn't consider the user scenario. Everything started to fall apart and this is the moment when it's too late to think about emergency access. If you are not in such a situation, let's think about the best way to organize secure access to the cluster.

First of all, we need authentication instead of the usual certificates. Take dexidp or keycloak and connect it to lDAP, Github or Google. Configure groups, scopes and claims and create RBAC Role and RoleBindings that will correspond to our groups. Now in a dev environment, users can go to the cluster and see how the applications are doing.

For production, this is no longer enough, we don't want anyone to be able to access the cluster at any time. Security experts are getting paranoid: What if an admin adds a new group to ldap? What if the account gets hijacked? What if an employee is bribed?

At this point, access negotiation becomes necessary. Additional participants are needed to grant access to the admin to do the work.

## Concept
Admin can request access for a certain period of time and send it for approval.

![](/docs/img/ar.png)

The responsible persons may grant access, but only if a quorum is present. 

![](/docs/img/coordination.png)

## Demo
```
cd demo/
./demo.sh
kubectl apply -f four-eyes.yml -f etcd.yml
```

```
kubectl --user oidc cluster-info
```
