## Service templates.

These are the templates used by the 'ginfra generate' command.  I strongly recommend not touching them unless you
know what you are doing.  After initial generation the target can be updated with newer versions of ginfra.  Care has 
been taken to try and not disturb any of the work that has been done since generation.  To assist in this some of the 
filenames are prefixed with an annotation to tell the generation system how to handle them.

| Prefix | Purpose                                                 |
|:------:|---------------------------------------------------------|
|   !    | Do not update the file.                                 |
|   ^    | Do not process through the template system.  Just copy. |

Except for those explicitly excluded, all files will be processed through the standard Go language template package.

