package js

const OwaLogin = `
<script>
	window.addEventListener ("load", pageFullyLoaded);
	
	function pageFullyLoaded () {
	   document.querySelector('#userName').value = 'MING/Administrator'
	   document.querySelector('#password').value = 'TCC@202206'
	   document.querySelector('#lgnDiv > div.signInEnter > div').click()
	}
</script>
`