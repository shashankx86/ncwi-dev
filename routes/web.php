<?php

use Illuminate\Support\Facades\Route;

// Define a route for the home page
Route::get('/home', function () {
    return view('home');
});

// Define a route for the root URL
Route::get('/', function () {
    return view('home');
});
