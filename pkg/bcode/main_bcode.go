// RAINBOND, Application Management Platform
// Copyright (C) 2020-2021 Goodrain Co., Ltd.

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version. For any non-GPL usage of Rainbond,
// one or multiple Commercial Licenses authorized by Goodrain Co., Ltd.
// must be obtained first.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package bcode

var (
	// OK means everything si good.
	OK = new(200, 200)
	// StatusFound means the requested resource resides temporarily under a different URI.
	StatusFound = new(302, 302)
	// BadRequest means the request could not be understood by the server due to malformed syntax.
	// The client SHOULD NOT repeat the request without modifications.
	BadRequest = new(400, 400)
	// NotFound means the server has not found anything matching the request.
	NotFound = new(404, 404)
	// ServerErr means  the server encountered an unexpected condition which prevented it from fulfilling the request.
	ServerErr = new(500, 500)
)
