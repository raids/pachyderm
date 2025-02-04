# Delete Data

If *bad* data was committed into a Pachyderm input repository, your
pipeline might result in an error. In this case, you might need to
delete this data to resolve the issue. Depending on the nature of
the bad data and whether or not the bad data is in the HEAD of
the branch, you can perform one of the following actions:

- [Delete the HEAD of a Branch](#delete-the-head-of-a-branch).
If the incorrect data was added in the latest commit and no additional
data was committed since then, follow the steps in this section to fix
the HEAD of the corrupted branch.
- [Delete Old Commits](#delete-old-commits). If after
committing the incorrect data, you have added more data to the same
branch, follow the steps in this section to delete corrupted files.


## Delete the HEAD of a Branch

To fix a broken HEAD, run the following command:

```shell
pachctl delete commit <repo>@<branch-or-commit-id>
```

When you delete a bad commit, Pachyderm performs the following actions:

- Deletes the commit metadata.
- Changes HEADs of all the branches that had the bad commit as their
  HEAD to the bad commit's parent. If the bad commit does not have
  a parent, Pachyderm sets the branch's HEAD to `nil`.
- If the bad commit has children, sets their parents to the deleted commit
  parent. If the deleted commit does not have a parent, then the
  children commit parents are set to `nil`.
- Deletes all the jobs that were triggered by the bad commit. Also,
  Pachyderm interrupts all running jobs, including not only the
  jobs that use the bad commit as a direct input but also the ones farther
  downstream in your DAG.
- Deletes the output commits from the deleted jobs. All the actions
  listed above are applied to those commits as well.

## Delete Old Commits

If you have committed more data to the branch after the bad data
was added, you can try to delete the commit as described in
[Delete the HEAD of a Branch](#delete-the-head-of-a-branch).
However, unless the subsequent commits overwrote or deleted the
bad data, the bad data might still be present in the
children commits. Deleting a commit does not modify its children.

In Git terms, `pachctl delete commit` is equivalent to squashing a
commit out of existence, such as with the `git reset --hard` command.
The `delete commit` command is not equivalent to reverting a
commit in Git. The reason for this
behavior is that the semantics of revert can get ambiguous
when the files that are being reverted have been
otherwise modified. Because Pachyderm is a centralized system
and the volume of data that you typically store in Pachyderm is
large, merge conflicts can quickly become untenable. Therefore,
Pachyderm prevents merge conflicts entirely.

To resolve issues with the commits that are not at the tip of the
branch, you can try to delete the children commits. However,
those commits might also have the data that you might want to
keep.

To delete a file in an older commit:

1. Start a new commit:

      ```shell
      pachctl start commit <repo>@<branch>
      ```

1. Delete all corrupted files from the newly opened commit:

      ```shell
      pachctl delete file <repo>@<branch or commitID>:/path/to/files
      ```

1. Finish the commit:

      ```shell
      pachctl finish commit <repo>@<branch>
      ```

4. Delete the initial bad commit and all its children up to
   the newly finished commit.

      Depending on how you use Pachyderm, the final step might be
      optional. After you finish the commit, the HEADs of all your
      branches converge to correct results as downstream jobs finish.
      However, deleting those commits cleans up your
      commit history and ensures that the errant data is not
      available when non-HEAD versions of the data is read.

