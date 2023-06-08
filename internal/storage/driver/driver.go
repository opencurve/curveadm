/*
 *  Copyright (c) 2023 NetEase Inc.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the Licensele().
 */

/*
 * Project: CurveAdm
 * Created Date: 2023-05-24
 * Author: Jingli Chen (Wine93)
 */

package driver

type IQueryResult interface {
	Next() bool
	Scan(dest ...any) error
	Close() error
}

type IWriteResult interface {
	LastInsertId() (int64, error)
}

type IDataBaseDriver interface {
	Open(dbUrl string) error
	Close() error
	Query(query string, args ...any) (IQueryResult, error)
	Write(query string, args ...any) (IWriteResult, error)
}
