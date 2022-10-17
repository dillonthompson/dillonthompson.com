import React, { FC } from "react"
import Image from "next/image"
import styles from './Header.module.css'
import img from '@/assets/images/dillon.png'

const Header: FC = () => {
  return (
    <div className={styles.header}>
      <div className={styles.logoContainer}>
        <div className={styles.logo}>
          <Image src={img} alt="dillon thompson" />
        </div>
      </div>
      <div className={styles.nav}>
        <div className={styles.items}>
          <div className={styles.item}>item</div>
          <div className={styles.item}>item</div>
          <div className={styles.item}>item</div>
          <div className={styles.item}>item</div>
        </div>
      </div>
    </div>
  )
}

export default Header