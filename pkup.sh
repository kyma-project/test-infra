#!/bin/sh
SINCE=${1:-$(perl -MPOSIX -le 'print strftime "%Y-%m-20", localtime(time - 60*60*24*31)')}
AUTHOR=mwisnia97@gmail.com
REPO=$(basename `git config remote.origin.url` ".git")
OUTPUT=pkup_$REPO\_$SINCE
echo ================================================================
echo $AUTHOR, $REPO, since $SINCE
echo ================================================================
rm -rf -- $OUTPUT
mkdir $OUTPUT
i=0
for sha1 in $(git rev-list HEAD --since=$SINCE --author=$AUTHOR --no-merges) ; do
	git format-patch -1 $sha1 -o $OUTPUT --start-number $i
	((++i))
done
tar -czf $OUTPUT.tar.gz $OUTPUT
rm -rf -- $OUTPUT
echo ================================================================
