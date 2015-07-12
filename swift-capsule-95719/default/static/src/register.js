function RegisterCtrl($scope, $http) {
	
	$scope.register = function() {
		
		var new_auth = {
			username: $scope.username,
			password: $scope.password,
			confirm_password: $scope.confirm_password,
			email: $scope.email
		};
		
		//console.log(new_auth);
		
		$http.post('/register', new_auth).
		success(function(data, status, headers, config) {
			
			$scope.register_status = "Registration successful!";
		}).
		error(function(data, status, headers, config) {
			
			$scope.register_status = "Registration failed: " + data;
		});
	};
}

app.controller('RegisterCtrl', RegisterCtrl);