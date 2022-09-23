import React, { FC } from "react"
import styles from './Header.module.css'

const Header: FC = () => {
  return (
    <header className={styles.header}>
      <nav className={styles.nav}>
        <ul className={styles.items}>
          <li>
            <div className={styles.image}>
              soon to be an image
            </div>
          </li>
          <li>
            <div className={styles.image}>
              soon to be a button
            </div>
          </li>
          <li>
            <div className={styles.image}>
              soon to be a button
            </div>
          </li>
          <li>
            <div className={styles.image}>
              soon to be a button
            </div>
          </li>
        </ul>
      </nav>
    </header>
  )
}

export default Header