# What is Cellar

Cellar is an application for sharing secrets in text form for a set period of time.

Cellar is very simple to use.
Enter your text and set an expiration.
Cellar will then generate an ID for your secret and securely encrypt and store it.
After the expiration you have set, the secret is automatically deleted.

# Why choose Cellar?

## How is Cellar different?

Cellar was inspired by the popular [Private Bin][priv-bin] but has a few key differences.

The primary difference is that Cellar integrates with existing systems to securely handle secrets.
For example, it does not handle any encryption itself.
Rather, it relies on an existing encryption as a service platform to encrypt your text.
That way, you can 
Cellar also doesn't do any sort of polling for expiration.
Instead, it publishes each secret with a set expiration and relies on the datastore to remove expired data automatically.

## When should I use Cellar?

Cellar is a quick way to share secrets without the need to authenticate or manage access control.
It is a great option if you want quickly share information with someone without the need to create any kind of account.

For example, when your friend needs your address for a wedding invitation.
Or maybe your family in another country needs a code to pickup something you sent them.
Perhaps, your boss needs the password to a zip file you sent them.

These are all great uses of Cellar.
Because Cellar relies on data expiration, access count limits, and randomly generated IDs to restrict access,
you know that your data stayed safe in transit and is only available as long as you want it to be.

## What doesn't Cellar do?

Cellar is not intended to replace proper password manager or authenticated secret sharing practices.  
If you are looking to permanently share passwords or files, a password manager or a cloud provider may be a better option.

Cellar uses long, randomly generated IDs to access data without authentication or authorization.  
That means that anyone with the ID that can access your installation of Cellar can read your data.
So make sure to set proper expirations and access limits.

If my data is only hidden behind an ID, couldn't some just guess the ID?  
Theoretically, yes.
Since there is no authentication, anyone who has the right ID can get to the password.
_However_, the IDs are 32 bytes of randomly generated data.
That means there are approximately 6.334028666297328e+49 different IDs.


[priv-bin]: https://github.com/PrivateBin/PrivateBin