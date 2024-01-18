helm package chart
mv core-* ../chartrepo
cd ../chartrepo
helm repo index --url  https://no8ge.github.io/chartrepo/ .
git add .
git commit -m 'update'
git push